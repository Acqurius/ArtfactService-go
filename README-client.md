# Artifact Client - JavaScript 使用說明

這是一個簡單易用的 JavaScript 客戶端，讓前端開發者可以輕鬆使用 token-based 的檔案上傳和下載功能。

## 📦 檔案說明

- **`artifact-client.js`** - 核心客戶端程式庫
- **`example-usage.html`** - 完整的使用範例（包含拖曳上傳、進度條等）

## 🚀 快速開始

### 1. 引入客戶端

```html
<script src="artifact-client.js"></script>
```

### 2. 初始化客戶端

```javascript
const client = new ArtifactClient('http://localhost:8080');
```

### 3. 上傳檔案

```javascript
// 從 file input 取得檔案
const file = document.getElementById('fileInput').files[0];

// 上傳檔案
const result = await client.uploadFile(file, {
  maxUploads: 5,
  onProgress: (percent) => {
    console.log(`上傳進度: ${percent}%`);
  }
});

console.log('檔案 UUID:', result.uuid);
console.log('檔案名稱:', result.filename);
```

### 4. 下載檔案

```javascript
// 使用 UUID 下載檔案
await client.downloadFile('artifact-uuid-here', 'myfile.pdf', {
  maxDownloads: 3
});
```

## 📖 API 文件

### `new ArtifactClient(baseUrl)`

建立客戶端實例。

**參數:**
- `baseUrl` (string) - API 伺服器的基礎 URL

**範例:**
```javascript
const client = new ArtifactClient('http://localhost:8080');
```

---

### `uploadFile(file, options)`

上傳檔案到 S3。

**參數:**
- `file` (File) - 要上傳的檔案物件
- `options` (Object) - 選項
  - `maxUploads` (number) - 最大上傳次數，預設 1
  - `validFrom` (string) - Token 生效時間 (ISO 8601)
  - `validTo` (string) - Token 過期時間 (ISO 8601)
  - `allowedCIDR` (string) - IP 限制 (CIDR 格式)
  - `onProgress` (function) - 進度回調函數 `(percent) => void`

**回傳:**
```javascript
{
  uuid: 'artifact-uuid',
  filename: 'example.pdf',
  size: 1024000,
  contentType: 'application/pdf',
  token: 'upload-token'
}
```

**範例:**
```javascript
const result = await client.uploadFile(file, {
  maxUploads: 5,
  validTo: '2026-12-31T23:59:59Z',
  allowedCIDR: '192.168.1.0/24',
  onProgress: (percent) => {
    progressBar.style.width = percent + '%';
  }
});
```

---

### `downloadFile(artifactUuid, filename, options)`

下載檔案。

**參數:**
- `artifactUuid` (string) - Artifact 的 UUID
- `filename` (string, optional) - 下載後的檔案名稱
- `options` (Object) - 選項
  - `maxDownloads` (number) - 最大下載次數，預設 1
  - `validFrom` (string) - Token 生效時間 (ISO 8601)
  - `validTo` (string) - Token 過期時間 (ISO 8601)
  - `allowedCIDR` (string) - IP 限制 (CIDR 格式)

**範例:**
```javascript
await client.downloadFile(
  'artifact-uuid-here',
  'downloaded-file.pdf',
  {
    maxDownloads: 3,
    validTo: '2026-12-31T23:59:59Z'
  }
);
```

---

### `getArtifactMetadata(artifactUuid)`

取得檔案的 metadata。

**參數:**
- `artifactUuid` (string) - Artifact 的 UUID

**回傳:**
```javascript
{
  uuid: 'artifact-uuid',
  filename: 'example.pdf',
  size: 1024000,
  content_type: 'application/pdf',
  uploaded_at: '2026-01-09T00:00:00Z'
}
```

## 💡 完整範例

### 基本上傳

```html
<input type="file" id="fileInput">
<button onclick="upload()">上傳</button>

<script src="artifact-client.js"></script>
<script>
  const client = new ArtifactClient('http://localhost:8080');
  
  async function upload() {
    const file = document.getElementById('fileInput').files[0];
    if (!file) {
      alert('請選擇檔案');
      return;
    }
    
    try {
      const result = await client.uploadFile(file);
      alert('上傳成功！UUID: ' + result.uuid);
    } catch (error) {
      alert('上傳失敗: ' + error.message);
    }
  }
</script>
```

### 帶進度條的上傳

```html
<input type="file" id="fileInput">
<div id="progress" style="width: 0%; height: 20px; background: blue;"></div>
<button onclick="uploadWithProgress()">上傳</button>

<script src="artifact-client.js"></script>
<script>
  const client = new ArtifactClient('http://localhost:8080');
  
  async function uploadWithProgress() {
    const file = document.getElementById('fileInput').files[0];
    const progressBar = document.getElementById('progress');
    
    try {
      const result = await client.uploadFile(file, {
        onProgress: (percent) => {
          progressBar.style.width = percent + '%';
          progressBar.textContent = percent + '%';
        }
      });
      
      alert('上傳成功！UUID: ' + result.uuid);
    } catch (error) {
      alert('上傳失敗: ' + error.message);
    }
  }
</script>
```

### 拖曳上傳

```html
<div id="dropZone" style="border: 2px dashed #ccc; padding: 50px;">
  拖曳檔案到這裡
</div>

<script src="artifact-client.js"></script>
<script>
  const client = new ArtifactClient('http://localhost:8080');
  const dropZone = document.getElementById('dropZone');
  
  dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.style.background = '#eee';
  });
  
  dropZone.addEventListener('dragleave', () => {
    dropZone.style.background = '';
  });
  
  dropZone.addEventListener('drop', async (e) => {
    e.preventDefault();
    dropZone.style.background = '';
    
    const file = e.dataTransfer.files[0];
    if (file) {
      try {
        const result = await client.uploadFile(file);
        alert('上傳成功！UUID: ' + result.uuid);
      } catch (error) {
        alert('上傳失敗: ' + error.message);
      }
    }
  });
</script>
```

## 🎨 查看完整範例

開啟 `example-usage.html` 查看完整的互動式範例，包含：
- ✅ 拖曳上傳
- ✅ 上傳進度條
- ✅ 下載功能
- ✅ 最近上傳的檔案列表
- ✅ 漂亮的 UI 設計

## 🔧 技術細節

### 上傳流程
1. 呼叫 `POST /genUploadPresignedURL` 取得 upload token
2. 呼叫 `POST /artifacts/upload/:token` 取得 S3 presigned URL
3. 直接 PUT 檔案到 S3（不經過 application server）

### 下載流程
1. 呼叫 `POST /genDownloadPresignedURL` 取得 download token
2. 呼叫 `GET /artifacts/:token` 取得檔案（Server 會 302 redirect 到 S3）
3. 瀏覽器自動跟隨 redirect 從 S3 下載

## 📝 注意事項

1. **CORS 設定**: 如果前端和後端在不同 domain，需要設定 CORS
2. **檔案大小限制**: 依照你的 S3 和 Server 設定
3. **Token 過期時間**: Presigned URL 預設 15 分鐘過期
4. **瀏覽器相容性**: 使用現代瀏覽器（支援 Fetch API 和 File API）

## 🚀 部署建議

### 在 React 中使用

```javascript
import ArtifactClient from './artifact-client.js';

function UploadComponent() {
  const client = new ArtifactClient('http://localhost:8080');
  const [progress, setProgress] = useState(0);
  
  const handleUpload = async (file) => {
    const result = await client.uploadFile(file, {
      onProgress: setProgress
    });
    console.log('Uploaded:', result.uuid);
  };
  
  return (
    <input type="file" onChange={(e) => handleUpload(e.target.files[0])} />
  );
}
```

### 在 Vue 中使用

```javascript
import ArtifactClient from './artifact-client.js';

export default {
  data() {
    return {
      client: new ArtifactClient('http://localhost:8080'),
      progress: 0
    };
  },
  methods: {
    async handleUpload(file) {
      const result = await this.client.uploadFile(file, {
        onProgress: (percent) => {
          this.progress = percent;
        }
      });
      console.log('Uploaded:', result.uuid);
    }
  }
};
```

## 📞 支援

如有問題請參考 `example-usage.html` 的完整範例。
