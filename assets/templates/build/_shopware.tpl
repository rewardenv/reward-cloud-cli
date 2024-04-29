{{- define "shopware-dockerfile-start" -}}
FROM rewardenv/php-fpm:{{ include "php-version" . }}-shopware-web{{ include "build-image-suffix" . }}

ARG COMPOSER_VERSION=2.4.4
{{- end -}}

{{- define "shopware-dockerfile-end" -}}
COPY --chown=www-data:www-data . .

RUN <<-EOF
    set -eux
    if [ "${COMPOSER_VERSION:-2}" = "1" ]; then
      alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer1"
    else
      alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
      if [ "${COMPOSER_VERSION:-2}" = "2" ]; then
        composer self-update --stable;
      else
        composer self-update "${COMPOSER_VERSION:-2}";
      fi
    fi
    if [ -n "${COMPOSER_GITHUB_USER-}" ]; then composer global config http-basic.github.com ${COMPOSER_GITHUB_USER-} ${COMPOSER_GITHUB_TOKEN-}; fi
    if [ -n "${COMPOSER_BITBUCKET_PUBLIC_KEY-}" ]; then composer global config bitbucket-oauth.bitbucket.org ${COMPOSER_BITBUCKET_PUBLIC_KEY-} ${COMPOSER_BITBUCKET_PRIVATE_KEY-}; fi
    if [ -n "${COMPOSER_GITLAB_OAUTH_TOKEN-}" ]; then composer global config gitlab-oauth.${COMPOSER_GITLAB_HOST:-gitlab.com} ${COMPOSER_GITLAB_OAUTH_TOKEN-}; fi
    if [ -n "${COMPOSER_GITLAB_TOKEN-}" ]; then composer global config gitlab-oauth.${COMPOSER_GITLAB_HOST:-gitlab.com} ${COMPOSER_GITLAB_TOKEN-}; fi
    if [ -n "${COMMAND_BEFORE_COMPOSER_INSTALL-}" ]; then eval "${COMMAND_BEFORE_COMPOSER_INSTALL-}"; fi
    if [ "${COMPOSER_INSTALL:-true}" = "true" ]; then composer install --no-interaction --no-progress; fi
    if [ -n "${COMMAND_AFTER_COMPOSER_INSTALL-}" ]; then eval "${COMMAND_AFTER_COMPOSER_INSTALL-}"; fi
    composer --no-interaction clear-cache
    if [ "${COMPOSER_DUMP_AUTOLOAD-}" = "true" ]; then composer dump-autoload -o; fi
    printf "<?php\nprintf(\"php-version: %%g </br>\", phpversion());\nprintf(\"build-date: $(date '+%Y/%m/%d %H:%M:%S')\");\n?>\n" > public/version.php
EOF
{{- end -}}
