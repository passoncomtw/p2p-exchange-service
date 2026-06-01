declare module '@pkg/logger' {
  export enum LogLevel {
    DEBUG = 0,
    INFO = 1,
    WARN = 2,
    ERROR = 3,
    NONE = 4,
  }

  interface Logger {
    debug: (message: string, ...args: any[]) => void;
    info: (message: string, ...args: any[]) => void;
    warn: (message: string, ...args: any[]) => void;
    error: (message: string, ...args: any[]) => void;
    setLevel: (level: LogLevel) => void;
    configure: (config: Partial<{
      level: LogLevel;
      enableTimestamp: boolean;
      enableColors: boolean;
    }>) => void;
    reset: () => void;
    getConfig: () => {
      level: LogLevel;
      enableTimestamp: boolean;
      enableColors: boolean;
    };
  }

  export const logger: Logger;
  export default logger;
}

