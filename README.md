# go-modular-template

Khuôn dự án backend **Go + [chi](https://github.com/go-chi/chi)** (phụ thuộc duy nhất là router; phần còn lại dùng thư viện chuẩn), được xếp theo một tư duy cấu trúc thư mục rõ ràng:

1. **Hạ tầng tách khỏi nghiệp vụ**: `framework/` và `modules/`. Phụ thuộc chỉ đi một chiều.
2. **Hạ tầng xếp theo năng lực, rồi mới tới nhà cung cấp**: `config/env`, `pubsub/memory`… Thêm Vault hay Kafka chỉ là thêm một thư mục anh em.
3. **Mỗi module là một lát dọc**: `dto / handler / service / repo / model`. Có module mẫu `hello` để copy.
4. **Tên file cho biết tầng và use case**: `routes/0_init.go`, `routes/<module>.go`.
5. **Mỗi loại code một chỗ**: code viết tay, code sinh tự động, tích hợp bên ngoài.

Chạy được ngay: có API mẫu, event trong process, log JSON, tắt êm, test, CI và Dockerfile.

---

## Chạy nhanh

```bash
make run                              # AUTH_TOKEN mặc định cho local: dev-only-token-change-me
curl localhost:8080/healthz
curl -X POST localhost:8080/hello -H 'Authorization: Bearer dev-only-token-change-me' -d '{"name":"An"}'
curl localhost:8080/hello/<id>
make check                            # gofmt + go vet + go test -race (cổng trước khi push)
```

| Biến | Mặc định | Ghi chú |
|---|---|---|
| `APP_ENV` | `local` | gắn vào mọi dòng log |
| `HTTP_ADDR` | `:8080` | |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `AUTH_TOKEN` | (bắt buộc) | ≥ 16 ký tự. Thiếu hoặc sai thì app **dừng ngay lúc khởi động** và báo hết các khoá có vấn đề |

---

## Cấu trúc thư mục

```
go-modular-template/
├── main.go                         # điểm vào: config → logger → server.New → Run; bắt SIGINT/SIGTERM
├── server/
│   └── server.go                   # NƠI LẮP RÁP: dựng hạ tầng, dựng từng module, nối route
├── routes/
│   ├── 0_init.go                   # chi.Router: middleware toàn cục, healthz, 404 JSON, mount từng module
│   ├── hello.go                    # route module hello dưới /hello — auth gắn THEO TỪNG ROUTE: r.With(auth)
│   └── routes_test.go              # test cả luồng qua HTTP; mọi route ghi không token → 401
├── middleware/
│   ├── middleware.go               # Chain, RequestID, Recover, AccessLog
│   └── auth.go                     # BearerToken (gắn cho từng route)
├── define/
│   └── define.go                   # hằng số dùng chung: tên service, topic, event
│
├── framework/                      # ★ HẠ TẦNG — không chứa nghiệp vụ, không import modules/
│   ├── base/                       #   hợp đồng chung cho mọi module
│   │   ├── errs/errs.go            #     kiểu lỗi chuẩn: Kind, Key, Msg, Retriable
│   │   └── event/event.go          #     phong bì event: ID, Name, Key (giữ thứ tự), Data, Attempt
│   ├── config/
│   │   ├── config.go               #   struct Config có kiểu + Load(Source): validate 1 lần, báo hết lỗi
│   │   ├── config_test.go
│   │   ├── env/env.go              #   nhà cung cấp: biến môi trường
│   │   └── vault/                  #   (thêm khi cần) nhà cung cấp: Vault — cùng interface Source
│   ├── log/log.go                  #   slog JSON ra stdout
│   ├── db/                         #   (thêm khi cần) lưu BỀN: postgres/, mongo/
│   ├── dbmem/
│   │   └── memory/memory.go        #   lưu TẠM trong RAM (generic, an toàn đa goroutine)
│   ├── pubsub/
│   │   ├── pubsub.go               #   interface Publisher / Subscriber / Handler
│   │   ├── memory/                 #   nhà cung cấp: bus trong process (lỗi 1 event không giết consumer)
│   │   │   ├── memory.go
│   │   │   └── memory_test.go
│   │   └── kafka/                  #   (thêm khi cần) nhà cung cấp: Kafka — retry topic + DLQ
│   ├── service/
│   │   ├── httpx/                  #   loại server HTTP: timeout chuẩn, tắt êm, JSON + map lỗi → status
│   │   │   ├── server.go
│   │   │   └── respond.go
│   │   └── grpcx/                  #   (thêm khi cần) loại server gRPC
│   └── utils/id.go                 #   tiện ích không thuộc năng lực nào (NewID bằng crypto/rand)
│
├── modules/                        # ★ NGHIỆP VỤ — mỗi domain một thư mục
│   └── hello/                      #   MODULE MẪU: copy khi tạo module mới
│       ├── dto/dto.go              #     dữ liệu vào/ra API
│       ├── model/model.go          #     dữ liệu lưu kho (KHÁC dto, không lộ field nội bộ)
│       ├── handler/hello.go        #     mọi cửa vào: HTTP (Create, Get) + event (OnGreetingCreated)
│       ├── service/                #     nghiệp vụ, chỉ phụ thuộc interface Repo + Publisher
│       │   ├── hello.go
│       │   └── hello_test.go
│       └── repo/memory.go          #     truy cập dữ liệu; thêm postgres.go cạnh đây khi cần
│
├── .github/workflows/ci.yml        # CI: make check + build + govulncheck + docker build
├── Makefile  Dockerfile  .dockerignore  .gitignore  .env.example  go.mod  go.sum
```

Những dòng ghi **(thêm khi cần)** chưa có trong repo. Chúng chỉ đánh dấu **chỗ** dành cho phần đó, để khi cần thì ai cũng đặt vào cùng một nơi. Repo không chứa thư mục rỗng hay code dựng sẵn cho sau này.

### Vì sao xếp như vậy

| Quy tắc | Thể hiện ở đâu | Lợi ích |
|---|---|---|
| Hạ tầng tách khỏi nghiệp vụ | `framework/` và `modules/` | người viết nghiệp vụ chỉ làm trong `modules/` |
| Phụ thuộc một chiều | `modules/` import `framework/`, chiều ngược lại thì không | đổi hạ tầng không chạm nghiệp vụ |
| Năng lực rồi mới tới nhà cung cấp | `config/{env,vault}`, `pubsub/{memory,kafka}`, `db/{postgres,mongo}` | thêm nhà cung cấp = thêm thư mục anh em |
| Lưu bền tách khỏi lưu tạm | `db/` và `dbmem/` | nhìn tên là biết dữ liệu nào mất được |
| Hợp đồng chung ở tầng thấp nhất | `base/errs`, `base/event` | mọi module trả lỗi và phát event cùng một hình dạng |
| Runtime theo loại server | `service/httpx`, `service/grpcx` | đổi hoặc thêm giao thức mà không đụng nghiệp vụ |
| Module theo lát dọc | `modules/<domain>/{dto,model,handler,service,repo}` | bê cả thư mục đi là tách được thành microservice |
| `dto` khác `model` | `hello/dto` và `hello/model` | API không vô tình trả ra field nội bộ (token, secret…) |
| Một chỗ biết dữ liệu nằm đâu | `repo/` | service test được bằng repo giả |
| Lắp ráp ở một chỗ | `server/server.go` | mỗi module chỉ nhận đúng thứ nó cần, không có "god object" |
| Bản đồ API ở một thư mục | `routes/0_init.go` + mỗi module một file | `0_` luôn đứng đầu; auth gắn theo từng route nên không phụ thuộc thứ tự đăng ký |
| Router là chi, nhưng middleware vẫn là `func(http.Handler) http.Handler` | `middleware/` không import chi | đổi router không phải viết lại middleware; chi **panic khi `Use()` đặt sau route**, chặn luôn lỗi "tưởng route đã có auth" |

---

## Luồng một request

```
client
  └─► RequestID ─► Recover ─► AccessLog ─► chi.Router ─► [r.With(BearerToken) nếu là route ghi]
          └─► handler.Create   (đọc JSON, giới hạn 1MB, từ chối field lạ)
                └─► service.Create   (validate → repo.Create → event.New → Publisher.Publish)
                      └─► repo.Memory → dbmem/memory.Store
          ◄── httpx.JSON 201  hoặc  httpx.Error (errs.Kind → 400/401/404/500; lỗi 500 chỉ log, không lộ chi tiết)
```

## Luồng một event

```
service.Create ─► pubsub.Publisher.Publish(topic="hello", event{Name:"hello.created", Key: id})
                       └─► pubsub/memory.Bus (hàng đợi có giới hạn; đầy thì trả ErrQueueFull, không âm thầm bỏ)
                              └─► Bus.Run ─► handler.OnGreetingCreated
                                   (handler panic hoặc lỗi → log, vòng tiêu thụ VẪN chạy tiếp)
```

---

## Thêm một module mới

1. Copy `modules/hello/` thành `modules/<tên>/`, rồi đổi tên package và type.
2. `model/` chứa dữ liệu lưu kho. `dto/` chứa dữ liệu API. Đừng dùng chung một struct.
3. `service/` định nghĩa interface `Repo` mà nó cần, và viết test bằng repo hoặc publisher giả.
4. `repo/` là bản cài cho kho (RAM, Postgres…).
5. `handler/` chứa mọi cửa vào: HTTP và event.
6. `routes/<tên>.go`: đăng ký route. **Route ghi phải dùng `r.With(auth)`.** Mount trong `routes/0_init.go` bằng `r.Route("/<tên>", ...)`.
7. `server/server.go`: dựng module (repo → service → handler), và subscribe event nếu có.
8. Thêm hằng số topic và event vào `define/define.go`.
9. `make check`.

Mỗi use case lớn nên có **một file, cùng tên ở mọi tầng**, ví dụ `handler/order_create.go` ↔ `service/order_create.go`.

## Thêm một nhà cung cấp hạ tầng

Ví dụ thêm Kafka:
- Tạo `framework/pubsub/kafka/`, cài `pubsub.Publisher` và `pubsub.Subscriber`.
- Đổi đúng **một dòng** trong `server/server.go` (`membus.New` thành `kafka.New`).
- Service và handler không phải sửa gì.
- Với Kafka nên có:
  - producer idempotent, dùng `event.Key` làm partition key;
  - consumer có recover cho từng message;
  - topic retry có độ trễ, cộng với DLQ;
  - chưa chuyển được message sang topic retry hoặc DLQ thì **không Ack**.

Làm tương tự với `framework/config/vault/` (cài `config.Source`), `framework/db/postgres/`, `framework/service/grpcx/`.

Code sinh tự động (protobuf, ABI…) đặt ở một thư mục riêng, ví dụ `define/gen/` hoặc `gen/`, kèm `//go:generate`. Không trộn với code viết tay.

---

## Quy ước

- **Lỗi:** service trả `errs.Invalid / NotFound / Internal`. Chỉ có `httpx.Error` đổi lỗi sang HTTP status. `Msg` được trả ra API, `Err` (lỗi gốc) chỉ dùng để log.
- **Event:** tên `<domain>.<việc đã xảy ra>`, `Key` là id của thực thể (để giữ thứ tự), `Data` là JSON.
- **Log:** JSON ra stdout, gọi bằng field có tên. Không log header, body hay secret. Đẩy log đi đâu tiếp là việc của agent thu log.
- **Config:** mọi khoá đọc ở `config.Load`, có kiểu, thiếu hoặc sai thì dừng khởi động. Secret không bao giờ log và không bao giờ commit (`.env` đã nằm trong `.gitignore`).
- **Test:** đặt cạnh code (`*_test.go`). Mỗi route ghi có test "không token thì 401". Logic có goroutine chạy test với `-race`.
- **Tắt máy:** bắt SIGTERM và SIGINT, ngừng nhận request mới, chờ request đang chạy tối đa 20 giây.

## Lệnh

| Lệnh | Việc |
|---|---|
| `make run` | chạy local |
| `make test` / `make race` | test / test có race detector |
| `make vet` / `make fmt` | go vet / gofmt |
| `make check` | gofmt-check + vet + race: **cổng trước khi push** |
| `make build` | build ra `bin/app` (tĩnh, CGO_ENABLED=0) |
| `make docker` | build image distroless, chạy user nonroot |

## Phiên bản Go

- `go.mod` ghi `go 1.22` là **bản tối thiểu** để biên dịch được.
- **Build thật** (Dockerfile `ARG GO_VERSION`, CI `1.27.x`) luôn dùng bản Go **còn nhận vá bảo mật**, xem https://go.dev/dl. Go chỉ hỗ trợ 2 bản mới nhất; tại thời điểm viết là 1.27 và 1.26.
- Máy dev nên cài Go còn được hỗ trợ. Nếu muốn khoá cứng bản build, thêm dòng `toolchain go1.27.x` vào `go.mod`.
- Chạy với `GOTOOLCHAIN=local` thì Go sẽ không tự tải toolchain khác khi một dependency đòi bản Go cao hơn.

## Docker

- **Multi-stage:** stage `golang:${GO_VERSION}-alpine` build ra binary tĩnh, stage chạy là `distroless/static:nonroot` (không shell, không package manager, có sẵn CA certs và tzdata). Image cuối khoảng **16MB**.
- **Cache BuildKit** cho module và build cache: lần build sau chỉ mất vài giây.
- `--platform=$BUILDPLATFORM` cùng `TARGETOS`/`TARGETARCH`: biên dịch chéo khi làm image nhiều kiến trúc, ví dụ `docker buildx build --platform linux/amd64,linux/arm64 .`.
- `USER 65532:65532` là **UID dạng số**, để Kubernetes `runAsNonRoot` kiểm được.
- `ENTRYPOINT ["/app"]` dạng exec nên `docker stop` hay SIGTERM trên K8s đi thẳng tới app, và app tắt êm.
- Image không có `curl`, nên health check dùng **probe của nền tảng**, ví dụ K8s `httpGet: /healthz`.
- Cần build tái lập tuyệt đối thì pin base image bằng digest (`...@sha256:...`).

CI (`.github/workflows/ci.yml`) chạy `make check`, `make build`, **govulncheck** (quét lỗ hổng trong code thật sự được gọi tới) và `docker build`.

## Những chỗ cố ý đơn giản (`shortcut:` trong code)

| Hiện tại | Khi nào nâng cấp | Nâng lên |
|---|---|---|
| Auth bằng một token tĩnh | có user thật | JWT (kiểm alg, `exp`, `iss`) hoặc OIDC |
| Kho trong RAM | cần dữ liệu còn sau khi restart | `framework/db/postgres` + `repo/postgres.go` |
| Bus trong RAM, không retry/DLQ | chạy nhiều instance, hoặc không được mất event | `framework/pubsub/kafka` |
| Ghi kho rồi mới publish (có thể lệch nhau) | event bắt buộc phải đi kèm dữ liệu | Transactional Outbox |

Tìm tất cả: `grep -rn "shortcut:" .`

## Đổi tên module path

```bash
go mod edit -module github.com/<you>/<repo>
grep -rl 'github.com/luannt2002/go-modular-template' --include='*.go' . | xargs sed -i 's#github.com/luannt2002/go-modular-template#github.com/<you>/<repo>#g'
make check
```
