{{- define "magento-dockerfile-start" -}}
FROM rewardenv/php-fpm:{{ include "php-version" . }}-magento2-web{{ include "build-image-suffix" . }}

ARG COMPOSER_INSTALL
ARG COMPOSER_VERSION=2.2.19
ARG COMPOSER_AUTH
ARG COMPOSER_GITHUB_USER
ARG COMPOSER_GITHUB_TOKEN
ARG COMPOSER_BITBUCKET_PUBLIC_KEY
ARG COMPOSER_BITBUCKET_PRIVATE_KEY
ARG COMPOSER_GITLAB_HOST
ARG COMPOSER_GITLAB_OAUTH_TOKEN
ARG COMPOSER_GITLAB_TOKEN
ARG MAGENTO_PUBLIC_KEY
ARG MAGENTO_PRIVATE_KEY
ARG MAGENTO_LANGUAGES
ARG MAGENTO_DI_COMPILE=true
ARG MAGENTO_STATIC_CONTENT_DEPLOY=false
ARG MAGENTO_DEV_STATIC_SIGN=true
ARG MAGENTO_DEV_JS_MERGE=true
ARG MAGENTO_DEV_JS_MINIFY=true
ARG MAGENTO_DEV_JS_BUNDLING=false
ARG MAGENTO_DEV_CSS_MERGE=true
ARG MAGENTO_DEV_CSS_MINIFY=true
ARG MAGENTO_DEV_HTML_MINIFY=false
ARG COMMAND_BEFORE_COMPOSER_INSTALL
ARG COMMAND_AFTER_COMPOSER_INSTALL
ARG COMMAND_BEFORE_DI_COMPILE
ARG COMMAND_AFTER_DI_COMPILE
ARG COMPOSER_DUMP_AUTOLOAD=true
{{- end -}}

{{- define "magento-dockerfile-end" -}}
COPY --chown=www-data:www-data . .

RUN <<-EOF
    set -eux
    if [ "${COMPOSER_VERSION:-2}" = "1" ]; then
      alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer1";
    else
      alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2";
      if [ "${COMPOSER_VERSION:-2}" = "2" ]; then
        composer self-update --stable;
      else
        composer self-update "${COMPOSER_VERSION:-2}";
      fi
    fi
    if [ -n "${MAGENTO_PUBLIC_KEY-}" ]; then composer global config http-basic.repo.magento.com ${MAGENTO_PUBLIC_KEY-} ${MAGENTO_PRIVATE_KEY-}; fi
    if [ -n "${COMPOSER_GITHUB_USER-}" ]; then composer global config http-basic.github.com ${COMPOSER_GITHUB_USER-} ${COMPOSER_GITHUB_TOKEN-}; fi
    if [ -n "${COMPOSER_BITBUCKET_PUBLIC_KEY-}" ]; then composer global config bitbucket-oauth.bitbucket.org ${COMPOSER_BITBUCKET_PUBLIC_KEY-} ${COMPOSER_BITBUCKET_PRIVATE_KEY-}; fi
    if [ -n "${COMPOSER_GITLAB_OAUTH_TOKEN-}" ]; then composer global config gitlab-oauth.${COMPOSER_GITLAB_HOST:-gitlab.com} ${COMPOSER_GITLAB_OAUTH_TOKEN-}; fi
    if [ -n "${COMPOSER_GITLAB_TOKEN-}" ]; then composer global config gitlab-oauth.${COMPOSER_GITLAB_HOST:-gitlab.com} ${COMPOSER_GITLAB_TOKEN-}; fi
    if [ -n "${COMMAND_BEFORE_COMPOSER_INSTALL-}" ]; then eval "${COMMAND_BEFORE_COMPOSER_INSTALL-}"; fi
    if [ "${COMPOSER_VERSION:-2}" != "1" ]; then composer config --no-plugins allow-plugins.magento/* true; fi
    if [ "${COMPOSER_VERSION:-2}" != "1" ]; then composer config --no-plugins allow-plugins.laminas/laminas-dependency-plugin true; fi
    if [ "${COMPOSER_VERSION:-2}" != "1" ]; then composer config --no-plugins allow-plugins.dealerdirect/phpcodesniffer-composer-installer true; fi
    if [ "${COMPOSER_VERSION:-2}" != "1" ]; then composer config --no-plugins allow-plugins.cweagans/composer-patches true; fi
    if [ "${COMPOSER_INSTALL:-true}" = "true" ]; then composer install --no-interaction --no-progress; fi
    if [ -n "${COMMAND_AFTER_COMPOSER_INSTALL-}" ]; then eval "${COMMAND_AFTER_COMPOSER_INSTALL-}"; fi
    composer --no-interaction clear-cache
    if [ -n "${COMMAND_BEFORE_DI_COMPILE-}" ]; then eval "${COMMAND_BEFORE_DI_COMPILE}"; fi
    if [ ! -f app/etc/config.php ]; then
      php bin/magento module:enable --no-interaction --all;
    fi
    if [ "${MAGENTO_DI_COMPILE-}" = "true" ] && [ "${MAGENTO_DI_COMPILE_ON_DEMAND:-false}" != "true" ]; then
      php bin/magento setup:di:compile --no-interaction;
    fi
    if [ -n "${COMMAND_AFTER_DI_COMPILE-}" ]; then eval "${COMMAND_AFTER_DI_COMPILE-}"; fi
    if [ "${COMPOSER_DUMP_AUTOLOAD-}" = "true" ]; then composer dump-autoload -o; fi
    if [ "${MAGENTO_STATIC_CONTENT_DEPLOY-}" = "true" ] && [ "${MAGENTO_SCD_ON_DEMAND:-false}" != "true" ];
      then bin/magento setup:static-content:deploy --no-interaction --jobs=$(nproc) -fv ${MAGENTO_LANGUAGES:-};
    fi
    printf "<?php\nprintf(\"php-version: %%g </br>\", phpversion());\nprintf(\"build-date: $(date '+%Y/%m/%d %H:%M:%S')\");\n?>\n" > pub/version.php
EOF
{{- end -}}
