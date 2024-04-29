{{- define "shopware-pwa-dockerfile-start" -}}
ARG NODE_VERSION={{ include "node-version" . }}
ARG BASE_IMAGE=alpine
FROM node:${NODE_VERSION}-${BASE_IMAGE}
{{- end -}}

{{- define "shopware-pwa-dockerfile-end" -}}
USER node
WORKDIR /app

COPY --chown=node:node . .

RUN <<-EOF
    set -eux
    if [ -n "${COMMAND_BEFORE_NPM_INSTALL-}" ]; then eval "${COMMAND_BEFORE_NPM_INSTALL}"; fi
    npm install
    if [ -n "${COMMAND_AFTER_NPM_INSTALL-}" ]; then eval "${COMMAND_AFTER_NPM_INSTALL}"; fi
EOF

RUN <<-EOF
    set -eux
    if [ -n "${COMMAND_BEFORE_NPM_BUILD-}" ]; then eval "${COMMAND_BEFORE_NPM_BUILD}"; fi
    npm run build
    if [ -n "${COMMAND_AFTER_NPM_BUILD-}" ]; then eval "${COMMAND_AFTER_NPM_BUILD}"; fi
EOF

EXPOSE 3000

ARG NPM_RUN_COMMAND="run start"
ENV NPM_RUN_COMMAND=${NPM_RUN_COMMAND}

CMD [ "/bin/sh", "-c", "npm ${NPM_RUN_COMMAND}" ]
{{- end -}}
