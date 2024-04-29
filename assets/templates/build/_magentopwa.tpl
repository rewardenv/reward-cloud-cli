{{- define "magento-pwa-base-image" -}}
{{- if eq (include "node-version" .) "16" -}}
alpine3.18
{{- else if eq (include "node-version" .) "18" -}}
alpine3.18
{{- else if eq (include "node-version" .) "20" -}}
alpine3.18
{{- else -}}
alpine
{{- end -}}
{{- end -}}

{{- define "magento-pwa-dockerfile-start" -}}
ARG NODE_VERSION={{ include "node-version" . }}
ARG BASE_IMAGE={{ include "magento-pwa-base-image" . }}
FROM node:${NODE_VERSION}-${BASE_IMAGE} as build
{{- end -}}

{{- define "magento-pwa-dockerfile-end" -}}
RUN apk --no-cache --virtual add \
    python3 \
    make \
    g++ \
    yarn

USER node
WORKDIR /app

COPY --chown=node:node . .

ENV CI=true

RUN <<-EOF
    set -eux
    if [ -n "${COMMAND_BEFORE_YARN_INSTALL-}" ]; then eval "${COMMAND_BEFORE_YARN_INSTALL}"; fi
    yarn install --frozen-lockfile
    if [ -n "${COMMAND_AFTER_YARN_INSTALL-}" ]; then eval "${COMMAND_AFTER_YARN_INSTALL}"; fi
EOF

ENV NODE_ENV=production
ENV BABEL_KEEP_ATTRIBUTES=true

RUN <<-EOF
    set -eux
    if [ -n "${COMMAND_BEFORE_YARN_BUILD-}" ]; then eval "${COMMAND_BEFORE_YARN_BUILD}"; fi
    yarn run build
    if [ -n "${COMMAND_AFTER_YARN_BUILD-}" ]; then eval "${COMMAND_AFTER_YARN_BUILD}"; fi
EOF

# MULTI-STAGE BUILD
FROM node:${NODE_VERSION}-${BASE_IMAGE}

USER node
WORKDIR /app

COPY --from=build /app .

EXPOSE 8080

ARG YARN_RUN_COMMAND="stage:venia"
ENV YARN_RUN_COMMAND=${YARN_RUN_COMMAND}

ENV NODE_ENV=production
ENV STAGING_SERVER_HOST=0.0.0.0
ENV STAGING_SERVER_PORT=8080
ENV UPWARD_JS_UPWARD_PATH=upward.yml
ENV CUSTOM_ORIGIN_ENABLED=0
ENV UPWARD_JS_BIND_LOCAL=1
ENV UPWARD_JS_LOG_URL=1
ENV ENABLE_EXPRESS_SERVER_COMPRESSION=true
ENV MAGENTO_BACKEND_EDITION=AC
ENV MAGENTO_BACKEND_URL=redacted
ENV CHECKOUT_BRAINTREE_TOKEN=redacted
ENV GOOGLE_MAPS_API_KEY=redacted

CMD [ "/bin/sh", "-c", "yarn ${YARN_RUN_COMMAND}" ]

{{- end -}}
