/**@type {Config} */
var config = {
  serv: {
    listen: ":1122",
    cache: 16,
    static: "data/build",
    cors: {
      enable: true,
      allowOrigins: ["*"],
    },
  },
  // https://blog.amarea.cn/archives/netease-cloudmusic-history-version.html
  meta: [
    {
      type: "share",
      name: "netease-win",
      share: {
        link: "https://1drv.ms/f/s!AkzLL0B1QOXrgpdzsUjs-OcH47cJzg?e=QvFZbI",
      },
    },
    {
      type: "share",
      name: "netease-anr",
      share: {
        link: "https://1drv.ms/f/s!AkzLL0B1QOXrgpAd1NptJ4BijvMaAA?e=F9qqzg",
      },
    },
    {
      type: "share",
      name: "netease-mac",
      share: {
        link: "https://1drv.ms/f/s!AkzLL0B1QOXrgpVHyflAiWuJahFzvg?e=CDkdJX",
      },
    },
  ],
  root: {
    type: "mount",
    name: "/",
    mount: [
      {
        type: "ref",
        name: "Windows",
        ref: "netease-win",
      },
      {
        type: "ref",
        name: "Android",
        ref: "netease-anr",
      },
      {
        type: "ref",
        name: "Mac",
        ref: "netease-mac",
      },
    ],
  },
  site: [
    {
      type: "ref",
      name: "192.168.10.22:1122",
      ref: "netease-win",
    },
  ],
};
console.log(JSON.stringify(config));
