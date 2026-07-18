const { getDefaultConfig } = require('expo/metro-config');
const path = require('path');

const projectRoot = __dirname;
// 專案根目錄下的 shared 共用核心。
const sharedRoot = path.resolve(projectRoot, '../../shared');

const config = getDefaultConfig(projectRoot);

// 讓 Metro 監看 shared 目錄（位於 app 專案根目錄之外）。
config.watchFolders = [sharedRoot];

const pkgRoot = path.resolve(projectRoot, 'src/pkg');

const aliases = {
  '@pkg/logger': path.resolve(pkgRoot, 'logger'),
  '@pkg/notifications': path.resolve(pkgRoot, 'notifications'),
  '@pkg/utils/sagaHelpers': path.resolve(pkgRoot, 'utils/sagaHelpers'),
  '@pkg/utils': path.resolve(pkgRoot, 'utils'),
  '@shared': path.resolve(sharedRoot, 'src/index.ts'),
};

config.resolver.resolveRequest = (context, moduleName, platform) => {
  if (aliases[moduleName]) {
    return context.resolveRequest(context, aliases[moduleName], platform);
  }
  return context.resolveRequest(context, moduleName, platform);
};

module.exports = config;
