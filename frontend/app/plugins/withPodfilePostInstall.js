const { withDangerousMod, withXcodeProject } = require('expo/config-plugins');
const fs = require('fs');
const path = require('path');

function withPodfileFixups(config) {
  return withDangerousMod(config, [
    'ios',
    (config) => {
      const podfilePath = path.join(config.modRequest.platformProjectRoot, 'Podfile');
      let podfile = fs.readFileSync(podfilePath, 'utf-8');

      const snippet = `
    # Fix deployment target warnings for third-party pods
    installer.pods_project.targets.each do |target|
      target.build_configurations.each do |config|
        if config.build_settings['IPHONEOS_DEPLOYMENT_TARGET'].to_f < 15.1
          config.build_settings['IPHONEOS_DEPLOYMENT_TARGET'] = '15.1'
        end
      end
    end`;

      if (!podfile.includes('Fix deployment target warnings')) {
        podfile = podfile.replace(
          /post_install do \|installer\|/,
          `post_install do |installer|${snippet}`
        );
        fs.writeFileSync(podfilePath, podfile);
      }

      return config;
    },
  ]);
}

function withSuppressDuplicateLibWarning(config) {
  return withXcodeProject(config, (config) => {
    const project = config.modResults;
    const buildConfigs = project.pbxXCBuildConfigurationSection();

    for (const key in buildConfigs) {
      const buildConfig = buildConfigs[key];
      if (typeof buildConfig !== 'object' || !buildConfig.buildSettings) continue;

      const flags = buildConfig.buildSettings.OTHER_LDFLAGS;
      if (!flags) continue;

      const flagStr = '"-Xlinker"';
      const flagVal = '"-no_warn_duplicate_libraries"';

      if (Array.isArray(flags)) {
        if (!flags.includes(flagStr) || !flags.some(f => f === flagVal)) {
          flags.push(flagStr, flagVal);
        }
      }
    }

    return config;
  });
}

module.exports = function withPodfilePostInstall(config) {
  config = withPodfileFixups(config);
  config = withSuppressDuplicateLibWarning(config);
  return config;
};
