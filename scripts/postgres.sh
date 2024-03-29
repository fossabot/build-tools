#!/usr/bin/env bash

# Creates a user for a service in a postgres database (postgresql) running inside kubernetes
# A secret with the credentials will also be created.
# Intended for local setup.
# Parameters
# 1 - the service name for which to create the database user, username and password will be the same as well
# 2 - the environment to use to access Kubernetes
postgres:create_database_user() {
  local SERVICE_NAME="${1}"
  local ENVIRONMENT="${2}"
  local KUBECTL_CMD=$(kubecmd ${ENVIRONMENT})
  local db_pod_name=$(${KUBECTL_CMD} get pods --selector 'app=postgresql' --output jsonpath={.items..metadata.name})
  local db_cmd="CREATE USER \\\"${SERVICE_NAME}\\\" WITH PASSWORD '${SERVICE_NAME}';CREATE DATABASE \\\"${SERVICE_NAME}\\\" WITH OWNER \\\"${SERVICE_NAME}\\\" ENCODING utf8"
  ${KUBECTL_CMD} exec -it ${db_pod_name} -- bash -c " echo \"${db_cmd}\" | psql -U postgres -f -"

  local SECRET_NAME="${SERVICE_NAME}-db"
  ${KUBECTL_CMD} delete secret ${SECRET_NAME} &> /dev/null || true
  ${KUBECTL_CMD} create secret generic ${SECRET_NAME} \
 --from-literal=POSTGRES_USERNAME="${SERVICE_NAME}" \
 --from-literal=POSTGRES_PASSWORD="${SERVICE_NAME}" \
 --from-literal=POSTGRES_DBNAME="${SERVICE_NAME}"
}
