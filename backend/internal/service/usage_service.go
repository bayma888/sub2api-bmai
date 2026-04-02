package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

var (
	ErrUsageLogNotFound = infraerrors.NotFound("USAGE_LOG_NOT_FOUND", "usage log not found")
)

// CreateUsageLogRequest 创建使用日志请求
type CreateUsageLogRequest struct {
	UserID                int64   `json:"user_id"`
	APIKeyID              int64   `json:"api_key_id"`
	AccountID             int64   `json:"account_id"`
	RequestID             string  `json:"request_id"`
	Model                 string  `json:"model"`
	InputTokens           int     `json:"input_tokens"`
	OutputTokens          int     `json:"output_tokens"`
	CacheCreationTokens   int     `json:"cache_creation_tokens"`
	CacheReadTokens       int     `json:"cache_read_tokens"`
	CacheCreation5mTokens int     `json:"cache_creation_5m_tokens"`
	CacheCreation1hTokens int     `json:"cache_creation_1h_tokens"`
	InputCost             float64 `json:"input_cost"`
	OutputCost            float64 `json:"output_cost"`
	CacheCreationCost     float64 `json:"cache_creation_cost"`
	CacheReadCost         float64 `json:"cache_read_cost"`
	TotalCost             float64 `json:"total_cost"`
	ActualCost            float64 `json:"actual_cost"`
	RateMultiplier        float64 `json:"rate_multiplier"`
	Stream                bool    `json:"stream"`
	DurationMs            *int    `json:"duration_ms"`
}

// UsageStats 使用统计
type UsageStats struct {
	TotalRequests     int64   `json:"total_requests"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalCacheTokens  int64   `json:"total_cache_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	TotalActualCost   float64 `json:"total_actual_cost"`
	AverageDurationMs float64 `json:"average_duration_ms"`
}

// UsageService 使用统计服务
type UsageService struct {
	usageRepo            UsageLogRepository
	userRepo             UserRepository
	entClient            *dbent.Client
	authCacheInvalidator APIKeyAuthCacheInvalidator
}

// NewUsageService 创建使用统计服务实例
func NewUsageService(usageRepo UsageLogRepository, userRepo UserRepository, entClient *dbent.Client, authCacheInvalidator APIKeyAuthCacheInvalidator) *UsageService {
	return &UsageService{
		usageRepo:            usageRepo,
		userRepo:             userRepo,
		entClient:            entClient,
		authCacheInvalidator: authCacheInvalidator,
	}
}

// Create 创建使用日志
func (s *UsageService) Create(ctx context.Context, req CreateUsageLogRequest) (*UsageLog, error) {
	// 使用数据库事务保证「使用日志插入」与「扣费」的原子性，避免重复扣费或漏扣风险。
	tx, err := s.entClient.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := ctx
	if err == nil {
		defer func() { _ = tx.Rollback() }()
		txCtx = dbent.NewTxContext(ctx, tx)
	}

	// 验证用户存在
	_, err = s.userRepo.GetByID(txCtx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 创建使用日志
	usageLog := &UsageLog{
		UserID:                req.UserID,
		APIKeyID:              req.APIKeyID,
		AccountID:             req.AccountID,
		RequestID:             req.RequestID,
		Model:                 req.Model,
		InputTokens:           req.InputTokens,
		OutputTokens:          req.OutputTokens,
		CacheCreationTokens:   req.CacheCreationTokens,
		CacheReadTokens:       req.CacheReadTokens,
		CacheCreation5mTokens: req.CacheCreation5mTokens,
		CacheCreation1hTokens: req.CacheCreation1hTokens,
		InputCost:             req.InputCost,
		OutputCost:            req.OutputCost,
		CacheCreationCost:     req.CacheCreationCost,
		CacheReadCost:         req.CacheReadCost,
		TotalCost:             req.TotalCost,
		ActualCost:            req.ActualCost,
		RateMultiplier:        req.RateMultiplier,
		Stream:                req.Stream,
		DurationMs:            req.DurationMs,
	}

	inserted, err := s.usageRepo.Create(txCtx, usageLog)
	if err != nil {
		return nil, fmt.Errorf("create usage log: %w", err)
	}

	// 扣除用户余额
	balanceUpdated := false
	if inserted && req.ActualCost > 0 {
		if err := s.userRepo.UpdateBalance(txCtx, req.UserID, -req.ActualCost); err != nil {
			return nil, fmt.Errorf("update user balance: %w", err)
		}
		balanceUpdated = true
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit transaction: %w", err)
		}
	}

	s.invalidateUsageCaches(ctx, req.UserID, balanceUpdated)

	return usageLog, nil
}

func (s *UsageService) invalidateUsageCaches(ctx context.Context, userID int64, balanceUpdated bool) {
	if !balanceUpdated || s.authCacheInvalidator == nil {
		return
	}
	s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
}

// GetByID 根据ID获取使用日志
func (s *UsageService) GetByID(ctx context.Context, id int64) (*UsageLog, error) {
	log, err := s.usageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get usage log: %w", err)
	}
	return log, nil
}

// ListByUser 获取用户的使用日志列表
func (s *UsageService) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, pagination, err := s.usageRepo.ListByUser(ctx, userID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs: %w", err)
	}
	return logs, pagination, nil
}

// ListByAPIKey 获取API Key的使用日志列表
func (s *UsageService) ListByAPIKey(ctx context.Context, apiKeyID int64, params pagination.PaginationParams) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, pagination, err := s.usageRepo.ListByAPIKey(ctx, apiKeyID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs: %w", err)
	}
	return logs, pagination, nil
}

// ListByAccount 获取账号的使用日志列表
func (s *UsageService) ListByAccount(ctx context.Context, accountID int64, params pagination.PaginationParams) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, pagination, err := s.usageRepo.ListByAccount(ctx, accountID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs: %w", err)
	}
	return logs, pagination, nil
}

// GetStatsByUser 获取用户的使用统计
func (s *UsageService) GetStatsByUser(ctx context.Context, userID int64, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetUserStatsAggregated(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:     stats.TotalRequests,
		TotalInputTokens:  stats.TotalInputTokens,
		TotalOutputTokens: stats.TotalOutputTokens,
		TotalCacheTokens:  stats.TotalCacheTokens,
		TotalTokens:       stats.TotalTokens,
		TotalCost:         stats.TotalCost,
		TotalActualCost:   stats.TotalActualCost,
		AverageDurationMs: stats.AverageDurationMs,
	}, nil
}

// GetStatsByAPIKey 获取API Key的使用统计
func (s *UsageService) GetStatsByAPIKey(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetAPIKeyStatsAggregated(ctx, apiKeyID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get api key stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:     stats.TotalRequests,
		TotalInputTokens:  stats.TotalInputTokens,
		TotalOutputTokens: stats.TotalOutputTokens,
		TotalCacheTokens:  stats.TotalCacheTokens,
		TotalTokens:       stats.TotalTokens,
		TotalCost:         stats.TotalCost,
		TotalActualCost:   stats.TotalActualCost,
		AverageDurationMs: stats.AverageDurationMs,
	}, nil
}

// GetStatsByAccount 获取账号的使用统计
func (s *UsageService) GetStatsByAccount(ctx context.Context, accountID int64, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetAccountStatsAggregated(ctx, accountID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get account stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:     stats.TotalRequests,
		TotalInputTokens:  stats.TotalInputTokens,
		TotalOutputTokens: stats.TotalOutputTokens,
		TotalCacheTokens:  stats.TotalCacheTokens,
		TotalTokens:       stats.TotalTokens,
		TotalCost:         stats.TotalCost,
		TotalActualCost:   stats.TotalActualCost,
		AverageDurationMs: stats.AverageDurationMs,
	}, nil
}

// GetStatsByModel 获取模型的使用统计
func (s *UsageService) GetStatsByModel(ctx context.Context, modelName string, startTime, endTime time.Time) (*UsageStats, error) {
	stats, err := s.usageRepo.GetModelStatsAggregated(ctx, modelName, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get model stats: %w", err)
	}

	return &UsageStats{
		TotalRequests:     stats.TotalRequests,
		TotalInputTokens:  stats.TotalInputTokens,
		TotalOutputTokens: stats.TotalOutputTokens,
		TotalCacheTokens:  stats.TotalCacheTokens,
		TotalTokens:       stats.TotalTokens,
		TotalCost:         stats.TotalCost,
		TotalActualCost:   stats.TotalActualCost,
		AverageDurationMs: stats.AverageDurationMs,
	}, nil
}

// GetDailyStats 获取每日使用统计（最近N天）
func (s *UsageService) GetDailyStats(ctx context.Context, userID int64, days int) ([]map[string]any, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	stats, err := s.usageRepo.GetDailyStatsAggregated(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get daily stats: %w", err)
	}

	return stats, nil
}

// Delete 删除使用日志（管理员功能，谨慎使用）
func (s *UsageService) Delete(ctx context.Context, id int64) error {
	if err := s.usageRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete usage log: %w", err)
	}
	return nil
}

// GetUserDashboardStats returns per-user dashboard summary stats.
func (s *UsageService) GetUserDashboardStats(ctx context.Context, userID int64) (*usagestats.UserDashboardStats, error) {
	stats, err := s.usageRepo.GetUserDashboardStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user dashboard stats: %w", err)
	}
	return stats, nil
}

// GetAPIKeyDashboardStats returns dashboard summary stats filtered by API Key.
func (s *UsageService) GetAPIKeyDashboardStats(ctx context.Context, apiKeyID int64) (*usagestats.UserDashboardStats, error) {
	stats, err := s.usageRepo.GetAPIKeyDashboardStats(ctx, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("get api key dashboard stats: %w", err)
	}
	return stats, nil
}

// GetUserUsageTrendByUserID returns per-user usage trend.
func (s *UsageService) GetUserUsageTrendByUserID(ctx context.Context, userID int64, startTime, endTime time.Time, granularity string) ([]usagestats.TrendDataPoint, error) {
	trend, err := s.usageRepo.GetUserUsageTrendByUserID(ctx, userID, startTime, endTime, granularity)
	if err != nil {
		return nil, fmt.Errorf("get user usage trend: %w", err)
	}
	return trend, nil
}

// GetUserModelStats returns per-user model usage stats.
func (s *UsageService) GetUserModelStats(ctx context.Context, userID int64, startTime, endTime time.Time) ([]usagestats.ModelStat, error) {
	stats, err := s.usageRepo.GetUserModelStats(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get user model stats: %w", err)
	}
	return stats, nil
}

// GetAPIKeyModelStats returns per-model usage stats for a specific API Key.
func (s *UsageService) GetAPIKeyModelStats(ctx context.Context, apiKeyID int64, startTime, endTime time.Time) ([]usagestats.ModelStat, error) {
	stats, err := s.usageRepo.GetModelStatsWithFilters(ctx, startTime, endTime, 0, apiKeyID, 0, 0, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get api key model stats: %w", err)
	}
	return stats, nil
}

// GetBatchAPIKeyUsageStats returns today/total actual_cost for given api keys.
func (s *UsageService) GetBatchAPIKeyUsageStats(ctx context.Context, apiKeyIDs []int64, startTime, endTime time.Time) (map[int64]*usagestats.BatchAPIKeyUsageStats, error) {
	stats, err := s.usageRepo.GetBatchAPIKeyUsageStats(ctx, apiKeyIDs, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get batch api key usage stats: %w", err)
	}
	return stats, nil
}

// ListWithFilters lists usage logs with admin filters.
func (s *UsageService) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters usagestats.UsageLogFilters) ([]UsageLog, *pagination.PaginationResult, error) {
	logs, result, err := s.usageRepo.ListWithFilters(ctx, params, filters)
	if err != nil {
		return nil, nil, fmt.Errorf("list usage logs with filters: %w", err)
	}
	return logs, result, nil
}

// GetGlobalStats returns global usage stats for a time range.
func (s *UsageService) GetGlobalStats(ctx context.Context, startTime, endTime time.Time) (*usagestats.UsageStats, error) {
	stats, err := s.usageRepo.GetGlobalStats(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get global usage stats: %w", err)
	}
	return stats, nil
}

// GetStatsWithFilters returns usage stats with optional filters.
func (s *UsageService) GetStatsWithFilters(ctx context.Context, filters usagestats.UsageLogFilters) (*usagestats.UsageStats, error) {
	stats, err := s.usageRepo.GetStatsWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("get usage stats with filters: %w", err)
	}
	return stats, nil
}

// leaderboardTitles maps leaderboard type to the top-1 title/badge.
var leaderboardTitles = map[usagestats.LeaderboardType]string{
	usagestats.LeaderboardTypeCost:       "\U0001f451 消费之王",
	usagestats.LeaderboardTypeRecharge:   "\U0001f48e 钻石大佬",
	usagestats.LeaderboardTypeTokens:     "\U0001f525 Token 巨鲸",
	usagestats.LeaderboardTypeRequests:   "\u26a1 请求狂魔",
	usagestats.LeaderboardTypeActiveDays: "\U0001f4aa 勤劳之星",
}

// ======================== Leaderboard Cache ========================
// 进程内缓存，排行榜数据5分钟更新一次，避免每次请求都查数据库
// 注意：这里只缓存不含 my_rank 的榜单数据（因为 my_rank 是用户维度的）

const leaderboardCacheTTL = 30 * time.Minute

type leaderboardCacheEntry struct {
	items     []usagestats.LeaderboardEntry
	fetchedAt time.Time
}

var (
	leaderboardCache   sync.Map // key: "type:period" → *leaderboardCacheEntry
)

// GetLeaderboard returns the leaderboard for the given type and time period.
func (s *UsageService) GetLeaderboard(ctx context.Context, lbType usagestats.LeaderboardType, period usagestats.LeaderboardPeriod, userID int64, limit int) (*usagestats.LeaderboardResponse, error) {
	startTime, endTime := leaderboardPeriodToTimeRange(period)

	// 检查缓存：排行榜数据每5分钟刷新一次
	cacheKey := fmt.Sprintf("lb:%s:%s", lbType, period)
	var items []usagestats.LeaderboardEntry

	if cached, ok := leaderboardCache.Load(cacheKey); ok {
		entry := cached.(*leaderboardCacheEntry)
		if time.Since(entry.fetchedAt) < leaderboardCacheTTL {
			items = entry.items
		}
	}

	// 缓存未命中或已过期，查数据库
	if items == nil {
		var err error
		items, err = s.usageRepo.GetLeaderboard(ctx, lbType, startTime, endTime, limit)
		if err != nil {
			return nil, fmt.Errorf("get leaderboard: %w", err)
		}
		// 写入缓存
		leaderboardCache.Store(cacheKey, &leaderboardCacheEntry{
			items:     items,
			fetchedAt: time.Now(),
		})
	}

	// Assign titles
	topTitle := leaderboardTitles[lbType]
	for i := range items {
		switch items[i].Rank {
		case 1:
			items[i].Title = topTitle
		case 2, 3:
			items[i].Title = "\U0001f3c6 榜上有名"
		}
	}

	resp := &usagestats.LeaderboardResponse{
		Type:   lbType,
		Period: period,
		Items:  items,
	}

	// Get current user's rank (skip for anonymous/public access)
	if userID > 0 {
		myRank, err := s.usageRepo.GetUserLeaderboardRank(ctx, userID, lbType, startTime, endTime)
		if err != nil {
			// Non-fatal: log and continue without my_rank
			fmt.Printf("warn: failed to get user %d leaderboard rank: %v\n", userID, err)
		}
		resp.MyRank = myRank
	}

	return resp, nil
}

// leaderboardPeriodToTimeRange converts a period string to start/end times.
func leaderboardPeriodToTimeRange(period usagestats.LeaderboardPeriod) (time.Time, time.Time) {
	now := time.Now()
	switch period {
	case usagestats.LeaderboardPeriodToday:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	case usagestats.LeaderboardPeriodWeek:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6)
		return start, now
	case usagestats.LeaderboardPeriodMonth:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, -1, 0)
		return start, now
	case usagestats.LeaderboardPeriodAll:
		return time.Time{}, time.Time{} // zero times = no filter
	default:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	}
}
