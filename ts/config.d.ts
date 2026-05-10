type Config = {
  /**服务器 */
  serv: Serv;
  /**元数据 */
  meta: Provider[];
  /**根目录 */
  root: Provider & odpt;
  /**站点 */
  site?: Provider[]; //Record<string, Provider>;
};

type Serv = {
  /**监听地址 */
  listen: string;
  /**缓存大小（单位 MB） */
  cache: number;
  /**静态文件目录 */
  static: string;
  /**跨站配置 */
  cors?: {
    enable: boolean;
    allowOrigins: string[];
  };
};

type Provider = {
  /**类型 */
  type: string;
  /**名称 */
  name: string;
} & (
  | {
      type: "share";
      /**分享 */
      share: {
        /**链接 */
        link: string;
        /**[缓存] 访问令牌 */
        token?: string;
        /**[缓存] 过期时间 */
        expire?: number;
        /**[缓存] business */
        root?: string;
        /**[缓存] business */
        path?: string;
      };
    }
  | {
      type: "ref";
      /**引用 */
      ref: string;
    }
  | {
      type: "mount";
      /**挂载 */
      mount?: Provider[];
    }
);

type odpt = {
  /**路径保护 */
  odpt?: {
    /**匹配前缀（以 / 开头） */
    prefix: string;
    /**密码 */
    password: string;
  }[];
};
