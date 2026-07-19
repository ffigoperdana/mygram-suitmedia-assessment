pipeline {
    agent any

    options {
        timeout(time: 45, unit: 'MINUTES')
        disableConcurrentBuilds()
    }

    parameters {
        booleanParam(name: 'PUSH_IMAGES', defaultValue: true, description: 'Push GHCR images after build checks. Only the main branch publishes production images.')
        string(name: 'IMAGE_TAG', defaultValue: '', description: 'Optional image tag. Defaults to git SHA.')
        string(name: 'PUBLIC_API_BASE_URL', defaultValue: '', description: 'Frontend build-time API URL. Leave empty for same-origin /api proxy in production.')
        booleanParam(name: 'USE_SAME_ORIGIN_API', defaultValue: true, description: 'Use the frontend Nginx /api proxy instead of a separate API subdomain.')
        booleanParam(name: 'CAP_ENABLED', defaultValue: true, description: 'Enable Cap captcha in the production frontend build.')
        string(name: 'CAP_BASE_URL', defaultValue: 'https://cap.fgdev.tech', description: 'Frontend Cap captcha base URL.')
        string(name: 'CAP_SITE_KEY', defaultValue: '8d1607b07b', description: 'Frontend Cap captcha site key. Required when CAP_ENABLED=true.')
        booleanParam(name: 'CAP_REQUIRED_ON_LOGIN', defaultValue: true, description: 'Require Cap captcha on the login form.')
        string(name: 'GHCR_OWNER_REPO', defaultValue: 'ghcr.io/ffigoperdana/mygram', description: 'Required when PUSH_IMAGES=true, for example ghcr.io/owner/mygram.')
        booleanParam(name: 'DEPLOY_TO_COOLIFY', defaultValue: true, description: 'Trigger Coolify redeploy after pushing main images. Requires Jenkins secret text credential coolify-api-token.')
        string(name: 'COOLIFY_BASE_URL', defaultValue: 'http://127.0.0.1:8000', description: 'Coolify base URL from the Docker host network used by the deploy trigger.')
        string(name: 'COOLIFY_RESOURCE_UUID', defaultValue: 'elqs1vtmi6hw7afeevjj1vum', description: 'Coolify application/resource UUID for MyGram.')
    }

    environment {
        GO_VERSION_HINT = '1.26.x'
        NODE_VERSION_HINT = '20'
        BACKEND_LOCAL_IMAGE = 'mygram-api:jenkins'
        FRONTEND_LOCAL_IMAGE = 'mygram-web:jenkins'
        CI_JWT_SECRET = 'ci-jwt-secret-that-is-long-enough-for-mygram'
        DEFAULT_PUBLIC_API_BASE_URL = ''
        DEFAULT_CAP_ENABLED = 'true'
        DEFAULT_CAP_BASE_URL = 'https://cap.fgdev.tech'
        DEFAULT_CAP_SITE_KEY = '8d1607b07b'
        DEFAULT_CAP_REQUIRED_ON_LOGIN = 'true'
        DEFAULT_GHCR_OWNER_REPO = 'ghcr.io/ffigoperdana/mygram'
        DEFAULT_COOLIFY_BASE_URL = 'http://127.0.0.1:8000'
        DEFAULT_COOLIFY_RESOURCE_UUID = 'elqs1vtmi6hw7afeevjj1vum'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
                script {
                    env.GIT_SHORT_SHA = sh(script: 'git rev-parse --short=12 HEAD', returnStdout: true).trim()
                    env.EFFECTIVE_IMAGE_TAG = params.IMAGE_TAG?.trim() ? params.IMAGE_TAG.trim() : env.GIT_SHORT_SHA
                    env.EFFECTIVE_USE_SAME_ORIGIN_API = params.USE_SAME_ORIGIN_API == null ? 'true' : params.USE_SAME_ORIGIN_API.toString()
                    env.EFFECTIVE_PUBLIC_API_BASE_URL = env.EFFECTIVE_USE_SAME_ORIGIN_API == 'true' ? '' : (params.PUBLIC_API_BASE_URL?.trim() ?: env.DEFAULT_PUBLIC_API_BASE_URL)
                    env.EFFECTIVE_CAP_ENABLED = params.CAP_ENABLED == null ? env.DEFAULT_CAP_ENABLED : params.CAP_ENABLED.toString()
                    env.EFFECTIVE_CAP_BASE_URL = params.CAP_BASE_URL?.trim() ?: env.DEFAULT_CAP_BASE_URL
                    env.EFFECTIVE_CAP_SITE_KEY = params.CAP_SITE_KEY?.trim() ?: env.DEFAULT_CAP_SITE_KEY
                    env.EFFECTIVE_CAP_REQUIRED_ON_LOGIN = params.CAP_REQUIRED_ON_LOGIN == null ? env.DEFAULT_CAP_REQUIRED_ON_LOGIN : params.CAP_REQUIRED_ON_LOGIN.toString()
                    env.EFFECTIVE_GHCR_OWNER_REPO = params.GHCR_OWNER_REPO?.trim() ?: env.DEFAULT_GHCR_OWNER_REPO
                    env.EFFECTIVE_PUSH_IMAGES = params.PUSH_IMAGES == null ? 'true' : params.PUSH_IMAGES.toString()
                    env.EFFECTIVE_DEPLOY_TO_COOLIFY = params.DEPLOY_TO_COOLIFY == null ? 'true' : params.DEPLOY_TO_COOLIFY.toString()
                    env.EFFECTIVE_COOLIFY_BASE_URL = params.COOLIFY_BASE_URL?.trim() ?: env.DEFAULT_COOLIFY_BASE_URL
                    env.EFFECTIVE_COOLIFY_RESOURCE_UUID = params.COOLIFY_RESOURCE_UUID?.trim() ?: env.DEFAULT_COOLIFY_RESOURCE_UUID
                }
            }
        }

        stage('Tool Versions') {
            steps {
                sh 'go version'
                sh 'node --version'
                sh 'npm --version'
                sh 'docker version'
                sh 'docker compose version || docker-compose --version'
            }
        }

        stage('Backend Quality') {
            steps {
                sh 'go mod download'
                sh 'go mod verify'
                sh 'go vet ./...'
                sh '''
                    JWT_SECRET="${CI_JWT_SECRET}" \
                    GIN_MODE=test \
                    DB_HOST="${DB_HOST:-localhost}" \
                    DB_USER="${DB_USER:-postgres}" \
                    DB_PASSWORD="${DB_PASSWORD:-admin}" \
                    DB_NAME="${DB_NAME:-finalproject_test}" \
                    DB_PORT="${DB_PORT:-5432}" \
                    DB_SSLMODE="${DB_SSLMODE:-disable}" \
                    go test -count=1 ./...
                '''
            }
        }

        stage('Frontend Quality') {
            steps {
                dir('mygram-frontend') {
                    sh 'npm ci'
                    sh 'npm run typecheck'
                    sh 'npm run lint'
                    sh 'npm run test'
                    sh '''
                        if [ "${EFFECTIVE_CAP_ENABLED}" = "true" ] && [ -z "${EFFECTIVE_CAP_SITE_KEY}" ]; then
                          echo "CAP_SITE_KEY is required when CAP_ENABLED=true"
                          exit 1
                        fi

                        VITE_API_BASE_URL="${EFFECTIVE_PUBLIC_API_BASE_URL}" \
                        VITE_USE_SAME_ORIGIN_API="${EFFECTIVE_USE_SAME_ORIGIN_API}" \
                        VITE_CAP_ENABLED="${EFFECTIVE_CAP_ENABLED}" \
                        VITE_CAP_BASE_URL="${EFFECTIVE_CAP_BASE_URL}" \
                        VITE_CAP_SITE_KEY="${EFFECTIVE_CAP_SITE_KEY}" \
                        VITE_CAP_REQUIRED_ON_LOGIN="${EFFECTIVE_CAP_REQUIRED_ON_LOGIN}" \
                        npm run build
                    '''
                }
            }
        }

        stage('Docker Build Check') {
            steps {
                sh 'docker build -t "${BACKEND_LOCAL_IMAGE}" .'
                sh '''
                    if [ "${EFFECTIVE_CAP_ENABLED}" = "true" ] && [ -z "${EFFECTIVE_CAP_SITE_KEY}" ]; then
                      echo "CAP_SITE_KEY is required when CAP_ENABLED=true"
                      exit 1
                    fi

                    docker build \
                      --build-arg VITE_API_BASE_URL="${EFFECTIVE_PUBLIC_API_BASE_URL}" \
                      --build-arg VITE_USE_SAME_ORIGIN_API="${EFFECTIVE_USE_SAME_ORIGIN_API}" \
                      --build-arg VITE_CAP_ENABLED="${EFFECTIVE_CAP_ENABLED}" \
                      --build-arg VITE_CAP_BASE_URL="${EFFECTIVE_CAP_BASE_URL}" \
                      --build-arg VITE_CAP_SITE_KEY="${EFFECTIVE_CAP_SITE_KEY}" \
                      --build-arg VITE_CAP_REQUIRED_ON_LOGIN="${EFFECTIVE_CAP_REQUIRED_ON_LOGIN}" \
                      -t "${FRONTEND_LOCAL_IMAGE}" \
                      ./mygram-frontend
                '''
            }
        }

        stage('Compose Config Check') {
            steps {
                sh '''
                    if docker compose version >/dev/null 2>&1; then
                      COMPOSE="docker compose"
                    else
                      COMPOSE="docker-compose"
                    fi

                    BACKEND_IMAGE="${BACKEND_LOCAL_IMAGE}" \
                    FRONTEND_IMAGE="${FRONTEND_LOCAL_IMAGE}" \
                    DB_NAME=finalproject \
                    DB_USER=postgres \
                    DB_PASSWORD=ci-postgres-password \
                    JWT_SECRET="${CI_JWT_SECRET}" \
                    JWT_EXPIRATION_HOURS=24 \
                    CORS_ALLOWED_ORIGINS=https://mygram.example.com,https://docs.mygram.example.com \
                    PUBLIC_OPENAPI_ENABLED=true \
                    SWAGGER_UI_MODE=public \
                    CAP_ENABLED=false \
                    CAP_BASE_URL=https://cap.fgdev.tech \
                    CAP_SITE_KEY=jenkins-site-key \
                    CAP_SECRET_KEY=jenkins-secret-key \
                    CAP_REQUIRED_ON_LOGIN=true \
                    S3_ENDPOINT=https://s3.fgdev.tech \
                    S3_REGION=garage \
                    S3_BUCKET=fgdev-media \
                    S3_ACCESS_KEY_ID=jenkins-access-key \
                    S3_SECRET_ACCESS_KEY=jenkins-secret-key \
                    S3_FORCE_PATH_STYLE=true \
                    S3_UPLOAD_MAX_MB=4 \
                    $COMPOSE -f docker-compose.prod.yml config

                    JWT_SECRET="${CI_JWT_SECRET}" \
                    DB_PASSWORD=admin \
                    CORS_ALLOWED_ORIGINS=http://localhost:3000 \
                    PUBLIC_API_BASE_URL=http://localhost:8080 \
                    VITE_CAP_ENABLED=false \
                    S3_ENDPOINT="" \
                    S3_BUCKET="" \
                    S3_ACCESS_KEY_ID="" \
                    S3_SECRET_ACCESS_KEY="" \
                    $COMPOSE -f docker-compose.fullstack.yml config
                '''
            }
        }

        stage('Push Images') {
            when {
                allOf {
                    branch 'main'
                    expression { return env.EFFECTIVE_PUSH_IMAGES == 'true' }
                }
            }
            steps {
                script {
                    def imagePrefix = env.EFFECTIVE_GHCR_OWNER_REPO?.trim()
                    if (!imagePrefix) {
                        error('GHCR_OWNER_REPO is required when PUSH_IMAGES=true, for example ghcr.io/owner/mygram')
                    }
                    env.BACKEND_REMOTE_IMAGE = "${imagePrefix}-api:${env.EFFECTIVE_IMAGE_TAG}"
                    env.FRONTEND_REMOTE_IMAGE = "${imagePrefix}-web:${env.EFFECTIVE_IMAGE_TAG}"
                    env.BACKEND_MAIN_IMAGE = "${imagePrefix}-api:main"
                    env.FRONTEND_MAIN_IMAGE = "${imagePrefix}-web:main"
                }
                sh 'docker tag "${BACKEND_LOCAL_IMAGE}" "${BACKEND_REMOTE_IMAGE}"'
                sh 'docker tag "${FRONTEND_LOCAL_IMAGE}" "${FRONTEND_REMOTE_IMAGE}"'
                sh 'docker tag "${BACKEND_LOCAL_IMAGE}" "${BACKEND_MAIN_IMAGE}"'
                sh 'docker tag "${FRONTEND_LOCAL_IMAGE}" "${FRONTEND_MAIN_IMAGE}"'
                withCredentials([
                    usernamePassword(
                        credentialsId: 'ghcr',
                        usernameVariable: 'GHCR_USERNAME',
                        passwordVariable: 'GHCR_TOKEN'
                    )
                ]) {
                    sh '''
                        set +x
                        printf "%s" "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USERNAME" --password-stdin
                    '''
                    sh 'docker push "${BACKEND_REMOTE_IMAGE}"'
                    sh 'docker push "${FRONTEND_REMOTE_IMAGE}"'
                    sh 'docker push "${BACKEND_MAIN_IMAGE}"'
                    sh 'docker push "${FRONTEND_MAIN_IMAGE}"'
                }
            }
        }

        stage('Trigger Coolify') {
            when {
                allOf {
                    branch 'main'
                    expression { return env.EFFECTIVE_PUSH_IMAGES == 'true' }
                    expression { return env.EFFECTIVE_DEPLOY_TO_COOLIFY == 'true' }
                }
            }
            steps {
                withCredentials([
                    string(credentialsId: 'coolify-api-token', variable: 'COOLIFY_API_TOKEN')
                ]) {
                    sh '''
                        set +x
                        if [ -z "${EFFECTIVE_COOLIFY_BASE_URL}" ] || [ -z "${EFFECTIVE_COOLIFY_RESOURCE_UUID}" ]; then
                          echo "COOLIFY_BASE_URL and COOLIFY_RESOURCE_UUID are required when DEPLOY_TO_COOLIFY=true"
                          exit 1
                        fi

                        COOLIFY_URL="${EFFECTIVE_COOLIFY_BASE_URL%/}"
                        COOLIFY_DEPLOY_URL="${COOLIFY_URL}/api/v1/deploy?uuid=${EFFECTIVE_COOLIFY_RESOURCE_UUID}&force=false"
                        echo "Triggering Coolify deploy at ${COOLIFY_URL} for resource ${EFFECTIVE_COOLIFY_RESOURCE_UUID}"
                        printf 'header = "Authorization: Bearer %s"\\n' "$COOLIFY_API_TOKEN" | docker run --rm -i --network host curlimages/curl:8.11.1 \
                          --fail --silent --show-error --location --request GET \
                          --connect-timeout 10 \
                          --max-time 30 \
                          --config - \
                          "$COOLIFY_DEPLOY_URL"
                    '''
                }
            }
        }
    }

    post {
        success {
            echo "MyGram pipeline completed. Backend and frontend quality checks passed."
            echo "Effective image tag: ${env.EFFECTIVE_IMAGE_TAG}"
            echo "Mutable production tags: ${env.BACKEND_MAIN_IMAGE ?: 'not pushed'} and ${env.FRONTEND_MAIN_IMAGE ?: 'not pushed'}"
        }
        failure {
            echo 'MyGram pipeline failed. Review the failing stage before deploying this revision.'
        }
    }
}
