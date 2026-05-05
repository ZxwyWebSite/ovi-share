## ZxwyWebSite/OVI-Share

### 简介

- 参考 OneDrive-Vercel-Index 项目开发
- 之前用的对象存储取消免费计划，花三天时间把老项目翻新一下，支持挂载分享链接

### 注意

- 仅为分享目录提供 UI，请勿非法滥用，否则后果由使用者自行承担，与开发者无关。

### 特点

- 分享挂载，避免账号风控。
- （TODO）VFS，自定义挂载参数。
- 注：暂不支持挂载单个文件，请使用文件夹分享

### 使用

- 目前只有 OneDrive 分享（支持个人版与商业版）与 虚拟目录（用于挂载）
- 首次启动自动生成默认配置
- 另外需要下载 `assets.zip` 解压为 `data/build` 提供[前端文件](https://github.com/ovlist/frontend)

### 配置

&emsp;已支持 TypeScript 类型提示，VsCode 打开 `ts` 目录的 `config.js` 即可，之后运行 `node ts/config.js` 也可直接生成配置文件。

<details>
<summary>结构</summary>

```jsonc
{
  // 服务器
  "serv": {
    // 监听地址
    "listen": ":1122",
    // 缓存大小（MB）
    "cache": 16,
    // 静态文件
    "static": "data/build"
  },
  // 元数据，可供引用
  "meta": [
    {
      "type": "..."
    }
  ],
  // 根目录
  "root": {
    "type": "..."
  }
}
```

</details>
<details>
<summary>分享（share）</summary>

挂载 OneDrive 分享链接，自动检测个人版与商业版。

```jsonc
{
  "type": "share",
  "name": "p0",
  "share": {
    "link": "https://1drv.ms/f/c/..."
    // 保留备用，可能缓存令牌
  }
}
```

</details>
<details>
<summary>挂载（mount）</summary>

虚拟目录，可自定义名称。

```json
{
  "type": "mount",
  "name": "/",
  "mount": [
    {
      "type": "..."
    }
  ]
}
```

</details>
<details>
<summary>引用（ref）</summary>

引用 `Meta` 中的定义。

```jsonc
{
  "type": "ref",
  "name": "business",
  // Meta 段对应的 name
  "ref": "b0"
}
```

</details>

### 鸣谢

- [SpencerWooo/OneDrive-Vercel-Index](https://github.com/spencerwooo/onedrive-vercel-index)
