# BMAI Deployment & CI/CD

> Updated: 2026-03-29

## Production Branch
- **Branch**: `bmai-sub2api-my-production20260329`
- **Purpose**: 私有功能分支，不提 PR，只在自己 VPS 跑自编译二进制
- **Created from**: main (synced to upstream v0.1.105)

## Version Naming Convention
```
bmai-v{upstream_version}.{patch_number}
```
| Example | Meaning |
|---------|---------|
| `bmai-v0.1.105.1` | Based on upstream v0.1.105, 1st release |
| `bmai-v0.1.105.2` | Same upstream, 2nd patch (e.g., added leaderboard) |
| `bmai-v0.1.110.1` | Synced to upstream v0.1.110, 1st release |

## CI/CD Files (BMAI-specific, isolated from official)
| File | Purpose |
|------|---------|
| `.github/workflows/bmai-release.yml` | GitHub Actions: triggered by `bmai-v*` tags |
| `.goreleaser.bmai.yaml` | GoReleaser config: linux-amd64 only, no Docker |
| `deploy/install-bmai.sh` | One-line install script for VPS |

## Release Flow
```bash
# 1. Push code to production branch
git push origin bmai-sub2api-my-production20260329

# 2. Tag and push (triggers GitHub Actions)
git tag bmai-v0.1.105.2
git push origin bmai-v0.1.105.2

# 3. GitHub Actions automatically:
#    - Builds frontend (pnpm)
#    - Compiles Go binary (linux-amd64)
#    - Creates GitHub Release with tar.gz + checksums.txt

# 4. VPS one-line install/upgrade:
curl -sSL https://raw.githubusercontent.com/bayma888/sub2api-bmai/bmai-sub2api-my-production20260329/deploy/install-bmai.sh | sudo bash
```

## Upstream Sync Flow
```bash
# 1. Sync upstream to local main
git fetch upstream
git checkout main
git merge upstream/main
git push origin main

# 2. Merge into production branch
git checkout bmai-sub2api-my-production20260329
git merge main

# 3. Tag new version (use new upstream version number)
git tag bmai-v0.1.110.1
git push origin bmai-sub2api-my-production20260329 --tags
```

## VPS Details
| Item | Value |
|------|-------|
| IP | `69.63.216.162` |
| Port | `8080` |
| Service name | `bmai-sub2api` |
| Install dir | `/opt/bmai-sub2api/` |
| Config dir | `/etc/bmai-sub2api/` |
| Service user | `bmai-sub2api` |
| Systemd unit | `/etc/systemd/system/bmai-sub2api.service` |

## VPS Common Commands
```bash
sudo systemctl status bmai-sub2api     # 看状态
sudo systemctl restart bmai-sub2api    # 重启
sudo systemctl stop bmai-sub2api       # 停止
sudo journalctl -u bmai-sub2api -f     # 看日志
```

## Pitfall: GoReleaser Version in Archive Name
- GoReleaser `{{ .Version }}` keeps full tag name for non-standard tags
- Tag `bmai-v0.1.105.1` → archive: `sub2api_bmai-v0.1.105.1_linux_amd64.tar.gz`
- install-bmai.sh uses `${LATEST_VERSION}` directly (not stripped) to match

## Pitfall: GitHub Raw Cache
- `raw.githubusercontent.com` caches files ~5 min
- After pushing script fixes, use commit hash URL to bypass:
  `curl -sSL "https://raw.githubusercontent.com/bayma888/sub2api-bmai/{commit_hash}/deploy/install-bmai.sh" | bash`
- Or use jsdelivr: `https://cdn.jsdelivr.net/gh/bayma888/sub2api-bmai@{branch}/deploy/install-bmai.sh`
