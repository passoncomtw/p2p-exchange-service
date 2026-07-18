const { getDefaultConfig } = require('expo/metro-config');
const path = require('path');

const projectRoot = __dirname;
const frontendPkgRoot = path.resolve(projectRoot, '../pkg');

const config = getDefaultConfig(projectRoot);

config.watchFolders = [...(config.watchFolders || []), frontendPkgRoot];

const pkgRoot = path.resolve(projectRoot, 'src/pkg');

const aliases = {
  '@pkg/logger': path.resolve(pkgRoot, 'logger'),
  '@pkg/notifications': path.resolve(pkgRoot, 'notifications'),
  '@pkg/utils/sagaHelpers': path.resolve(pkgRoot, 'utils/sagaHelpers'),
  '@pkg/utils': path.resolve(pkgRoot, 'utils'),
  '@shared': path.resolve(projectRoot, 'src/shared'),
};

config.resolver.resolveRequest = (context, moduleName, platform) => {
  const resolved = aliases[moduleName];
  if (resolved) {
    const fs = require('fs');
    for (const ext of ['.ts', '.tsx', '.js', '.jsx']) {
      const indexFile = path.join(resolved, 'index' + ext);
      if (fs.existsSync(indexFile)) {
        return { type: 'sourceFile', filePath: indexFile };
      }
    }
    for (const ext of ['.ts', '.tsx', '.js', '.jsx']) {
      const filePath = resolved + ext;
      if (fs.existsSync(filePath)) {
        return { type: 'sourceFile', filePath };
      }
    }
    return context.resolveRequest(context, resolved, platform);
  }
  return context.resolveRequest(context, moduleName, platform);
};

module.exports = config;
