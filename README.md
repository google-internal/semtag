# semtag

語義化版本標籤工具——根據 Git commit 與分支自動推導版本號

## 功能特性

- 🚀 依循 Conventional Commits 自動找出下一個版本號
- 📌 取得當前分支可達的最大版本標籤
- 🌿 依據分支自動附加預發布尾綴（例如 beta、alpha、rc）
- 🔄 同步涵蓋穩定版與預發布版的版本策略

## 安裝

### 使用 `go install`

```bash
go install github.com/google-internal/semtag/cmd/semtag@latest
```

### 從 Release 下載二進位檔

官方 Release 會提供常用平台的單檔可執行檔，下載後加上執行權限即可使用：

```bash
# macOS (arm64)
curl -L https://github.com/google-internal/semtag/releases/latest/download/semtag_darwin_arm64 -o semtag
chmod +x semtag
sudo mv semtag /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/google-internal/semtag/releases/latest/download/semtag_linux_amd64 -o semtag
chmod +x semtag
sudo mv semtag /usr/local/bin/
```

### 從原始碼建置

```bash
git clone https://github.com/google-internal/semtag.git
cd semtag
go build -o semtag ./cmd/semtag
```

## 使用方法

### 命令總覽

`semtag` 預設執行 `next`（亦即 `semtag` 等同 `semtag next`）。目前支援的子命令如下：

| 命令 | 說明 |
|------|------|
| `next` | 根據 commit 與分支自動計算下一個版本號 |
| `current` | 取得當前分支的最新版本號（若無標籤則回傳 `0.0.0`） |

### 計算版本號

```bash
# 預設（無前綴）計算下一個語義化版本
semtag next

# 取得目前分支的最大版本號
semtag current
```

### 建立並推送標籤

搭配 `git` 指令即可建立並推送標籤：

```bash
# 計算含前綴的下一個版本
VERSION=$(semtag next -p v)

# 建立附註標籤
git tag -a "$VERSION" -m "Release ${VERSION}"

# 推送標籤
git push origin "refs/tags/${VERSION}"
```

指定 `-p`（或 `--prefix`）後，輸出會直接附帶該前綴，可立即拿來作為標籤名稱使用。

### 常用參數

- `--prefix`, `-p`：指定版本前綴（預設為空字串）
- `--branch-suffix`：僅對 `next` 生效，自訂分支對應的預發布尾綴，格式 `branch:suffix`，可重複指定，例如 `--branch-suffix develop:beta --branch-suffix release/*:rc`

### 協助指令

```bash
semtag --help
semtag next --help
```

## 版本計算規則

### Conventional Commits

`semtag` 依循 [Conventional Commits](https://www.conventionalcommits.org/) 規範決定版本號的升級類型：

#### Major 版本（x.0.0）

偵測到破壞性變更時會提升 Major 版本：

```bash
# commit 訊息包含 '!' 標記
git commit -m "feat!: 重大功能調整"
git commit -m "chore!: 破壞性調整"

# commit body 包含 BREAKING CHANGE
git commit -m "feat: 新功能" -m "BREAKING CHANGE: API 重大調整"
```

#### Minor 版本（0.x.0）

偵測到新增功能時會提升 Minor 版本：

```bash
git commit -m "feat: 新增匯出功能"
git commit -m "feat(auth): 實作使用者驗證"
```

#### Patch 版本（0.0.x）

偵測到錯誤修正時會提升 Patch 版本：

```bash
git commit -m "fix: 修正登入失敗"
git commit -m "fix(api): 修正 API 回應錯誤"
```

### 分支與預發布版本

`semtag` 會依據當前分支自動附加預發布尾綴：

| 分支名稱 | 尾綴 | 範例 |
|----------|------|------|
| `beta`, `develop`, `dev`, `staging` | `-beta.N` | `1.2.3-beta.1` |
| `alpha`, `experimental`, `wip` | `-alpha.N` | `1.2.3-alpha.1` |
| `rc`, `release`, `release/*` | `-rc.N` | `1.2.3-rc.1` |
| `hotfix`, `hotfix/*` | `-beta.N` | `1.2.3-beta.1` |
| `feature/*` | `-alpha.N` | `1.2.3-alpha.1` |
| `main`, `master` | 無尾綴 | `1.2.3` |

### 環境變數設定

可以透過環境變數覆寫分支與尾綴的對應關係：

```bash
# 方式一：使用逗號分隔的 key:value 列表
export SVU_BRANCH_SUFFIX_MAPPING="mybranch:custom,another:test"

# 方式二：個別設定環境變數
export SVU_BRANCH_MYBRANCH_SUFFIX="custom"
```

## 工作流程範例

### 範例一：主分支發佈

```bash
# 位於 main 分支
git checkout main

# 新增功能
git commit -m "feat: 新增匯出功能"

# 計算版本並建立標籤
VERSION=$(semtag next -p v)
git tag -a "$VERSION" -m "Release ${VERSION}"
git push origin "refs/tags/${VERSION}"
# 輸出：v1.3.0
```

### 範例二：開發分支預發布

```bash
git checkout develop
git commit -m "feat: 新的實驗功能"
semtag next
# 輸出：1.3.0-beta.1

git commit -m "fix: 修正 bug"
semtag next
# 輸出：1.3.0-beta.2
```

### 範例三：在 CI/CD 中使用 CLI

```yaml
- name: 計算版本號
  id: version
  run: |
    VERSION=$(semtag next -p v)
    echo "version=$VERSION" >> "$GITHUB_OUTPUT"

- name: 建立並推送標籤
  run: |
    VERSION="${{ steps.version.outputs.version }}"
    git tag -a "$VERSION" -m "Release ${VERSION}"
    git push origin "refs/tags/${VERSION}"
```

## CI/CD 工作流

專案內建 `Release` 工作流（位於 `.github/workflows/release.yml`），當 `main` 分支有推送時會：

- 透過 `go run ./cmd/semtag` 計算下一個版本並建立標籤
- 針對 Linux amd64 與 macOS arm64 交叉編譯單一可執行檔
- 建立 GitHub Release 並上傳對應平台的二進位檔

## 專案結構

```
.
├── cmd/
│   └── semtag/         # CLI 入口與版本計算邏輯
│       ├── main.go
│       ├── branch_suffix.go
│       ├── version.go
│       └── version_test.go
├── go.mod               # Go 模組定義
├── internal/
│   └── git/            # Git 操作封裝
│       ├── git.go
│       └── git_test.go
├── Dockerfile          # Docker 映像建置
└── README.md           # 本文件
```

## 開發

### 執行測試

```bash
go test ./...
```

### 建置

```bash
go build -o semtag ./cmd/semtag
```

### 建置 Docker 映像

```bash
docker build -t semtag .
```

## License

詳見 LICENSE 檔案。

