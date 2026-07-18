const { getDefaultConfig } = require('expo/metro-config');
const path = require('path');

const projectRoot = __dirname;
const sharedRoot = path.resolve(projectRoot, '../../shared');
const frontendPkgRoot = path.resolve(projectRoot, '../pkg');

const config = getDefaultConfig(projectRoot);

config.watchFolders = [sharedRoot, frontendPkgRoot];

const pkgRoot = path.resolve(projectRoot, 'src/pkg');

const aliases = {
  '@frontend-pkg/notifications': path.resolve(frontendPkgRoot, 'notifications'),
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
