export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  NONE = 4,
}

interface LoggerConfig {
  level: LogLevel;
  enableTimestamp: boolean;
  enableColors: boolean;
}

const defaultConfig: LoggerConfig = {
  level: __DEV__ ? LogLevel.DEBUG : LogLevel.WARN,
  enableTimestamp: true,
  enableColors: true,
};

let config = { ...defaultConfig };

function shouldLog(level: LogLevel): boolean {
  return level >= config.level;
}

function formatMessage(level: string, message: string): string {
  if (config.enableTimestamp) {
    const now = new Date().toISOString();
    return `[${now}] [${level}] ${message}`;
  }
  return `[${level}] ${message}`;
}

const logger = {
  debug: (message: string, ...args: any[]) => {
    if (shouldLog(LogLevel.DEBUG)) {
      console.log(formatMessage('DEBUG', message), ...args);
    }
  },
  info: (message: string, ...args: any[]) => {
    if (shouldLog(LogLevel.INFO)) {
      console.info(formatMessage('INFO', message), ...args);
    }
  },
  warn: (message: string, ...args: any[]) => {
    if (shouldLog(LogLevel.WARN)) {
      console.warn(formatMessage('WARN', message), ...args);
    }
  },
  error: (message: string, ...args: any[]) => {
    if (shouldLog(LogLevel.ERROR)) {
      console.error(formatMessage('ERROR', message), ...args);
    }
  },
  setLevel: (level: LogLevel) => {
    config.level = level;
  },
  configure: (partial: Partial<LoggerConfig>) => {
    config = { ...config, ...partial };
  },
  reset: () => {
    config = { ...defaultConfig };
  },
  getConfig: () => ({ ...config }),
};

export { logger };
export default logger;
