#!/bin/sh

DEPLOYMENT_NAME="provider-test"
LOG_LEVEL="info"
ATHENZ_URL="athenz.io/zts/v1"
MAX_REPLICA=50
MIN_REPLICA=10
CPU_UTILIZATION=50
APP_IMAGE="nginx:latest"
APP_NAME="nginx"
SIDECAR_IMAGE="yahoojapan/authorization-proxy:latest"
SIDECAR_NAME="authorization-proxy"
APP_PORT=80
SIDECAR_PORT=8080
SIDECAR_CONFIG_FILE="config.yaml"
SIDECAR_CONFIG_PATH="/etc/athenz/provider"
SIDECAR_CONFIG_NAME="provider-config"
SIDECAR_SECRET_NAME="provider-secret"
INGRESS_HOST="k8s.athenz.sample.host"

YAML="provider.yaml"

rm -rf ${YAML}
cat <<-EOF > ${YAML}
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  annotations: {}
  labels:
    app: ${APP_NAME}
  name: ${DEPLOYMENT_NAME}
  namespace: default
spec:
  progressDeadlineSeconds: 600
  replicas: ${MIN_REPLICA}
  selector:
    matchLabels:
      app: ${APP_NAME}
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: ${APP_NAME}
    spec:
      containers:
      - image: ${APP_IMAGE}
        imagePullPolicy: Always
        name: ${APP_NAME}
        ports:
        - containerPort: ${APP_PORT}
          protocol: TCP
        resources: {}
      - args:
        - -f
        - ${SIDECAR_CONFIG_PATH}/${SIDECAR_CONFIG_FILE}
        image: ${SIDECAR_IMAGE}
        imagePullPolicy: Always
        name: ${SIDECAR_NAME}
        ports:
        - containerPort: ${SIDECAR_PORT}
          protocol: TCP
        readinessProbe:
          failureThreshold: 2
          httpGet:
            path: /healthz
            port: 6082
          initialDelaySeconds: 3
          periodSeconds: 3
          successThreshold: 1
          timeoutSeconds: 2
        resources: {}
        volumeMounts:
        - mountPath: ${SIDECAR_CONFIG_PATH}
          name: ${SIDECAR_CONFIG_NAME}-volume
        - mountPath: ${SIDECAR_CONFIG_PATH}/keys
          name: ${SIDECAR_SECRET_NAME}-volume
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
      volumes:
      - configMap:
          defaultMode: 420
          name: ${SIDECAR_CONFIG_NAME}
        name: ${SIDECAR_CONFIG_NAME}-volume
      - name: ${SIDECAR_SECRET_NAME}-volume
        secret:
          defaultMode: 420
          secretName: ${SIDECAR_SECRET_NAME}
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: ${APP_NAME}
  name: ${DEPLOYMENT_NAME}
  namespace: default
spec:
  externalTrafficPolicy: Cluster
  ports:
  - name: app
    port: ${APP_PORT}
    protocol: TCP
    targetPort: ${APP_PORT}
  - name: sidecar
    port: ${SIDECAR_PORT}
    protocol: TCP
    targetPort: ${SIDECAR_PORT}
  selector:
    app: ${APP_NAME}
  sessionAffinity: None
  type: NodePort
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  labels:
    app: ${DEPLOYMENT_NAME}
  name: ${DEPLOYMENT_NAME}
  namespace: default
spec:
  rules:
  - host: ${INGRESS_HOST}
    http:
      paths:
      - backend:
          serviceName: ${DEPLOYMENT_NAME}
          servicePort: ${SIDECAR_PORT}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${SIDECAR_CONFIG_NAME}
  namespace: default
data:
  ${SIDECAR_CONFIG_FILE}: |
    version: "v2.0.0"
    server:
      port: ${SIDECAR_PORT}
      timeout: 10s
      shutdownTimeout: 10s
      shutdownDelay: 9s
      tls:
        enable: false
        certPath: ""
        keyPath: ""
        caPath: ""
      healthCheck:
        port: 6082
        endpoint: /healthz
      debug:
        enable: false
        port: 6083
        dump: true
        profiling: true
    athenz:
      url: ${ATHENZ_URL}
      timeout: 30s
      caPath: ""
      # caPath: /etc/ssl/cert.pem
    proxy:
      scheme: http
      host: localhost
      port: ${APP_PORT}
      bufferSize: 4096
      originHealthCheckPaths: []
    authorization:
      athenzDomains:
      - provider-domain1
      - provider-domain2
      publicKey:
        sysAuthDomain: sys.auth
        refreshPeriod: 24h
        retryDelay: ""
        eTagExpiry: 168h
        eTagPurgePeriod: 84h
      policy:
        expiryMargin: 48h
        refreshPeriod: 1h
        purgePeriod: 24h
        retryDelay: ""
        retryAttempts: 0
      jwk:
        refreshPeriod: ""
        retryDelay: ""
      accessToken:
        enable: true
        verifyCertThumbprint: true
        verifyClientID: true
        authorizedClientIDs:
          common_name1:
            - client_id1
            - client_id2
          common_name2:
            - client_id1
            - client_id2
        certBackdateDuration: 1h
        certOffsetDuration: 1h
      roleToken:
        enable: true
        roleAuthHeader: Athenz-Role-Auth
    log:
      level: ${LOG_LEVEL}
      color: false
---
apiVersion: v1
kind: Secret
metadata:
  name: ${SIDECAR_SECRET_NAME}
data:
  rootca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVORENDQXh5Z0F3SUJBZ0lFQnlkY0pqQU5CZ2txaGtpRzl3MEJBUVVGQURCYU1Rc3dDUVlEVlFRR0V3SkoKUlRFU01CQUdBMVVFQ2hNSlFtRnNkR2x0YjNKbE1STXdFUVlEVlFRTEV3cERlV0psY2xSeWRYTjBNU0l3SUFZRApWUVFERXhsQ1lXeDBhVzF2Y21VZ1EzbGlaWEpVY25WemRDQlNiMjkwTUI0WERURXhNRGd4T0RFNE16WXpNMW9YCkRURTRNRGd3T1RFNE16VTBPVm93V2pFTE1Ba0dBMVVFQmhNQ1NsQXhJekFoQmdOVkJBb01Ha041WW1WeWRISjEKYzNRZ1NtRndZVzRnUTI4dUxDQk1kR1F1TVNZd0pBWURWUVFEREIxRGVXSmxjblJ5ZFhOMElFcGhjR0Z1SUZCMQpZbXhwWXlCRFFTQkhNakNDQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFMYmNkdnU1ClJQc1NmRlN3dTBGMWRQQTFSNTRudWtORVJXQVp6VVFLc25qbCtoNGtPd0lmYUhkZzlPc2lCUW8zYnR2M0ZTQzcKUFZQVTBCR08xT3RudnRqZEJUZVVRU1VqNzVvUW84UDNBTDI2SnBKbmdWQ3BUNTZSUEU0Z3VsSi8vMHhOanFxdApUbCs4SjVjQ0tmMlZnMG0vQ3JxeE5SZzFxWE9JWWxHc0ZCYzBVT2VmeHZPVFhibkZBRTgza0hxQkQ5VDFjaW5vCmpHS3NjVHZ6THQ4cVhPbSs1MVlrZ2lpYXZ6MzljVUw5eFh0ck53bEhVRDV5a2FvN3hVK2RFbTQ5Z0FOVVNVRVYKUFBLR1JIUW85Ym1qRzl0Mngrb0RpYUJnNlZIMm9XUStkSnZiS3NzWVBNSG5hQmlKN0tzNExsQzViMjRWTXlnZApMOVdBRjRZaTh4ME00SWNDQXdFQUFhT0NBUUF3Z2Ywd0VnWURWUjBUQVFIL0JBZ3dCZ0VCL3dJQkFEQlRCZ05WCkhTQUVUREJLTUVnR0NTc0dBUVFCc1Q0QkFEQTdNRGtHQ0NzR0FRVUZCd0lCRmkxb2RIUndPaTh2WTNsaVpYSjAKY25WemRDNXZiVzVwY205dmRDNWpiMjB2Y21Wd2IzTnBkRzl5ZVM1alptMHdEZ1lEVlIwUEFRSC9CQVFEQWdFRwpNQjhHQTFVZEl3UVlNQmFBRk9XZFdUQ0NSMWpNclBvSVZEYUdlenExQkUzd01FSUdBMVVkSHdRN01Ea3dONkExCm9ET0dNV2gwZEhBNkx5OWpaSEF4TG5CMVlteHBZeTEwY25WemRDNWpiMjB2UTFKTUwwOXRibWx5YjI5ME1qQXkKTlM1amNtd3dIUVlEVlIwT0JCWUVGQnZramU4NmNXc1NaV2pQdHBHOE9VTUJqWFhKTUEwR0NTcUdTSWIzRFFFQgpCUVVBQTRJQkFRQnRLKzNwajdZcDFyWXd1dVp0dGNOVDBzbTRDazVJbi9FL09pcTArM1NXNXIwWXZLZDV3SGpCCk9ib2c0MDZBMGlUVnBYdC9ZcVBhMUE4TnFaMnF4ZW04Q01sSVpwaWV3UG5lcTIzbHNEUENjTkNXMXg1dm1BUVYKWTBpN21vVmRHMm56dEUvenBuQVdEeUVaZjYyd0F6bEpob3lpYzA2VDNDRUJhTER2RFhBYWVxS3l6Q0pDa1ZTOQpySEFFalV4Yy9EcWlrdmI1S2hKQXpYYTNadlRYMHF2ZWppelozUWsxTnlkV0M2NjJycHFEWVBCZmYvQ3RzeHo2CnVIUmZ4K3pBRHEzWXc4K2YwakFPWEZFZlBobml3ZEtwa0EvbVY3bXZCSGFpOGdnRUpRbzF1M01FTWRDWVJuODIKd1dFV280cU1tZDRRQmZMZTdhVUpaSmVFajBLb2V5TEUKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
---
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  annotations: {}
  name: ${DEPLOYMENT_NAME}
  namespace: default
spec:
  maxReplicas: ${MAX_REPLICA}
  minReplicas: ${MIN_REPLICA}
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ${DEPLOYMENT_NAME}
  targetCPUUtilizationPercentage: ${CPU_UTILIZATION}
EOF
