const { getDefaultConfig } = require('expo/metro-config');
const path = require('path');

const projectRoot = __dirname;
const config = getDefaultConfig(projectRoot);

const pkgRoot = path.resolve(projectRoot, 'src/pkg');

const pkgAliases = {
  '@pkg/logger': path.resolve(pkgRoot, 'logger'),
  '@pkg/utils/sagaHelpers': path.resolve(pkgRoot, 'utils/sagaHelpers'),
  '@pkg/utils': path.resolve(pkgRoot, 'utils'),
};

config.resolver.resolveRequest = (context, moduleName, platform) => {
  if (pkgAliases[moduleName]) {
    return context.resolveRequest(context, pkgAliases[moduleName], platform);
  }
  return context.resolveRequest(context, moduleName, platform);
};

module.exports = config;
