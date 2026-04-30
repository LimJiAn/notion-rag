# notion-rag

Notion에 기록한 문서를 수집하고, 임베딩과 벡터 검색을 거쳐 답변하는 개인용 RAG 프로젝트입니다. 

## 개요

이 프로젝트는 아래 흐름으로 동작합니다.

1. Notion 루트 페이지부터 하위 페이지를 재귀 수집합니다.
2. 문서를 청킹하고 Gemini 임베딩을 생성합니다.
3. 임베딩 결과를 로컬 저장소에 보관합니다.
4. 질문이 들어오면 유사 문서를 검색합니다.
5. 검색 결과를 바탕으로 답변을 생성합니다.

## 기술 스택

- Backend: Go, Gin, Swagger
- Frontend: React, TypeScript, Vite, Tailwind CSS
- Runtime: Docker, Docker Compose
- LLM: Gemini
- Source: Notion API

## 폴더 구조

```text
.
├── backend
│   ├── cmd/server
│   ├── docs
│   └── internal
│       ├── app
│       ├── chunk
│       ├── clients
│       │   ├── gemini
│       │   └── notion
│       ├── config
│       ├── domain
│       ├── httpapi
│       │   ├── middleware
│       │   ├── router
│       │   └── v1
│       │       ├── handlers
│       │       ├── requests
│       │       └── responses
│       ├── repositories
│       ├── services
│       └── settings
├── frontend
│   └── src
│       ├── components
│       ├── features
│       ├── hooks
│       └── lib
├── docker-compose.yml
└── Makefile
```

## 주요 기능

- Notion 문서 재귀 수집
- 텍스트 청킹
- Gemini 임베딩 생성
- 로컬 파일 기반 벡터 저장
- 유사도 기반 검색
- 답변 생성
- 런타임 설정 조회/수정 API
- Swagger UI 제공
- React 기반 웹 UI 제공

## 빠른 시작

`.env` 파일 생성:

```bash
make env
```

필수 설정:

- `NOTION_TOKEN`
- `NOTION_ROOT_PAGE_IDS`
- `GEMINI_API_KEY`
- `VITE_API_BASE_URL`

`NOTION_ROOT_PAGE_IDS`에는 Notion 페이지 UUID를 넣어야 합니다.

예시:

```env
NOTION_ROOT_PAGE_IDS=2d105bf2-ff64-80c6-9fd9-eb2f39a07bc8
VITE_API_BASE_URL=http://localhost:8080
```

주의:

- 해당 페이지는 Notion integration에 공유되어 있어야 합니다.
- 여러 페이지를 시작점으로 쓸 경우 쉼표로 구분합니다.

전체 실행:

```bash
make up
```

중지:

```bash
make down
```

재시작:

```bash
make restart
```

## 접속 주소

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI JSON: `http://localhost:8080/swagger/doc.json`

프론트의 `Settings` 탭에서도 Notion/Gemini 설정을 수정할 수 있습니다.
프론트에서 백엔드를 호출하는 주소는 `VITE_API_BASE_URL`로 설정합니다.

## Makefile 커맨드

```bash
make help
make up
make down
make build
make logs
make ps
make restart
make sync
make query q="지난주에 내가 정리한 업무 내용을 요약해줘"
make test
make fmt
make swagger
make frontend-install
make frontend-ensure
make frontend-build
make frontend-dev
make backend-run
make backend-run BACKEND_PORT=8081
make compose-config
make clean
make clean-deps
make clean-data
```

## API 예시

헬스 체크:

```bash
curl http://localhost:8080/healthz
```

통계 조회:

```bash
curl http://localhost:8080/api/v1/stats
```

Notion 동기화:

```bash
curl -X POST http://localhost:8080/api/v1/sync
```

질문:

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H 'Content-Type: application/json' \
  -d '{"question":"지난주에 내가 정리한 업무 내용을 요약해줘"}'
```

설정 조회:

```bash
curl http://localhost:8080/api/v1/settings
```

설정 수정:

```bash
curl -X PUT http://localhost:8080/api/v1/settings \
  -H 'Content-Type: application/json' \
  -d '{
    "notion_token": "secret_xxx",
    "notion_version": "2026-03-11",
    "notion_root_page_ids": "2d105bf2-ff64-80c6-9fd9-eb2f39a07bc8",
    "gemini_api_key": "your-api-key",
    "embedding_model": "gemini-embedding-001",
    "generation_model": "gemini-2.5-flash"
  }'
```

`notion_root_page_ids`는 현재 API에서 문자열 CSV 형식으로 받습니다.

예시:

```json
"notion_root_page_ids": "uuid-1,uuid-2"
```

## 로컬 개발

백엔드만 실행:

```bash
make backend-run
```

로컬 실행 시 백엔드는 `.env`를 읽되, Docker 전용 경로인 `/app/data` 대신 프로젝트 루트의 `.local-data`를 사용합니다.
8080 포트가 이미 사용 중이면 아래처럼 포트를 바꿀 수 있습니다.
Docker Compose 서비스가 이미 떠 있다면 `make down`으로 내린 뒤 실행해도 됩니다.

```bash
make backend-run BACKEND_PORT=8081
```

프론트 개발 서버 실행:

```bash
make frontend-dev
```

`make frontend-dev`와 `make frontend-build`는 `node_modules/.bin/vite`가 없으면 `yarn install --frozen-lockfile`을 먼저 실행합니다.
의존성을 강제로 다시 설치하려면 `make frontend-install`을 사용합니다.

Swagger 문서 재생성:

```bash
make swagger
```

## 참고

- 현재 저장소는 로컬 파일 기반입니다.
- 향후 `pgvector` 같은 외부 벡터 저장소로 교체하기 쉽도록 계층을 분리해 두었습니다.
