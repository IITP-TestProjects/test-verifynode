# Verify Node Mock
CEF 연동 테스트를 위한 Committee Verification Mock Server
## 프로젝트 개요

이 repository는 IITP 과제 "노드 간 메시지 전달과 합의를 위한 최적 경로 네트워크 프로토콜 기술개발"의 committee election 테스트 구성을 위한 Verify Node mock 구현이다.

CEF 기반 committee election 테스트 구성에서 `Blockchain/Client Node - CEF Server - Verify Node` 흐름을 기준으로, 본 repository는 CEF middleware가 전달한 committee candidate 목록을 받아 `CommitteeInfo`를 반환하는 gRPC 서버 역할을 수행한다.

현재 구현은 실제 운영용 검증 노드라기보다 CEF 연동 테스트를 위한 단순 mock이다. 후보의 VRF proof, public key, seed를 검증하거나 committee 선정 알고리즘을 수행하지 않고, 요청에 포함된 후보 `node_id` 목록을 그대로 committee member 목록으로 반환한다.

## 전체 IITP Committee Election 테스트 구성에서의 역할

전체 테스트 시스템은 다음 세 구성요소로 나뉜다.

1. `CEF-server`
   - CEF Interface Server와 MongoDB Replica Set으로 구성된다.
   - blockchain/client node와 verify node 사이에서 gRPC 기반 committee election middleware 역할을 수행한다.
   - CEF gRPC server port는 `50051`이다.
   - MongoDB replica set은 `mongo1`, `mongo2`, `mongo3` / `rs0` 구성이다.
   - candidate/node history를 저장한다.

2. `test-verifier`
   - 이 repository이다.
   - CEF가 전달한 `CommitteeRequest`를 받는 verify node mock이다.
   - gRPC service는 `committee.CommitteeService`이다.
   - 주요 RPC는 `RequestCommittee(CommitteeRequest) returns (CommitteeInfo)`이다.
   - 테스트 구성에서 CEF는 보통 `-verifynode=verify-server1:50053` 형태로 이 서버에 접속한다.

3. `test-client`
   - blockchain/client node mock이다.
   - CEF의 `mesh.Mesh` gRPC service에 접속해 committee candidate 정보와 서명 관련 데이터를 전송하고, CEF가 broadcast하는 committee 결과를 수신한다.
   - 주요 RPC는 `JoinNetwork`, `RequestCommittee`, `RequestAggregatedCommit`이다.

본 repository는 위 흐름 중 CEF가 committee 후보 목록을 외부 verify node에 질의하는 지점을 모사한다. CEF의 candidate 수집, MongoDB 저장, aggregate public key/commit 생성, committee 결과 broadcast는 이 repository가 아니라 CEF/client mock 쪽 책임이다.

## 시스템 아키텍처

```text
Blockchain / Client Node Mock
        |
        | CEF mesh.Mesh gRPC
        v
CEF Middleware
        |
        | committee.CommitteeService.RequestCommittee
        | target: verify-server1:50053
        v
test-verifier
```

`test-verifier`는 gRPC 서버를 `:50053`에서 열고 `committee.CommitteeService`를 등록한다. CEF가 `CommitteeRequest`를 보내면 서버는 후보 목록을 읽어 `CommitteeInfo`를 반환한다.

## Repository 구조

```text
.
├── Dockerfile
├── README.md
├── docker-compose.yml
├── go.mod
├── go.sum
├── main.go
├── network.go
├── proto_verify/
│   ├── verify.pb.go
│   └── verify_grpc.pb.go
├── verifierDockerBuild.sh
└── verify.proto
```

주요 파일은 다음과 같다.

- `main.go`: gRPC 서버 생성, `CommitteeService` 등록, TCP `:50053` listen을 수행한다.
- `network.go`: `CommitteeService.RequestCommittee` RPC의 mock 응답 로직을 구현한다.
- `verify.proto`: verify node gRPC service와 protobuf message 정의이다.
- `proto_verify/`: `verify.proto`에서 생성된 Go protobuf/gRPC 코드이다.
- `Dockerfile`: Go binary를 빌드하고 runtime image에서 `verifier`를 실행한다.
- `docker-compose.yml`: `verify-server1` 컨테이너를 `50053:50053`으로 실행한다.
- `verifierDockerBuild.sh`: `verifier:<version>` Docker image를 빌드한다.

## 주요 컴포넌트

### gRPC server

`main.go`는 다음 순서로 서버를 시작한다.

1. `newVerifyService()`로 `verifySrv` 인스턴스를 생성한다.
2. `grpc.NewServer()`로 gRPC server를 생성한다.
3. `pv.RegisterCommitteeServiceServer(s, srv)`로 `committee.CommitteeService`를 등록한다.
4. `net.Listen("tcp", ":50053")`로 port `50053`을 연다.
5. `s.Serve(lis)`로 요청을 처리한다.

서버 시작 로그는 다음과 같다.

```text
Verify Service is running on port 50053
```

### CommitteeService 구현

`network.go`의 `RequestCommittee` 구현은 다음 동작을 수행한다.

1. `"Received committee request"` 로그를 출력한다.
2. `CommitteeRequest.Candidates`를 읽는다.
3. 각 candidate의 `NodeId`를 `member_ids` 목록에 추가한다.
4. 첫 번째 candidate의 `NodeId`를 `leader_member_id`로 사용한다.
5. 현재 Unix timestamp를 문자열로 변환해 `timestamp`에 넣는다.
6. 현재 mock 구현에서는 요청의 `channel` 값을 응답에 반영하지 않고, `channel_id`를 고정 문자열 `"<test>"`로 반환한다. 따라서 channel 단위 검증이 필요한 테스트에서는 이 부분을 수정해야 한다.

현재 코드 기준으로 candidate 검증/선정 기준은 구현되어 있지 않다. `seed`, `proof`, `publickey`, `addr`, `port` 필드는 protobuf message에는 존재하지만 `RequestCommittee` 로직에서는 사용되지 않는다.

## gRPC / protobuf 인터페이스

`verify.proto`의 package와 service는 다음과 같다.

```proto
package committee;

service CommitteeService {
  rpc RequestCommittee(CommitteeRequest) returns (CommitteeInfo) {}
}
```

생성된 gRPC full method name은 다음과 같다.

```text
/committee.CommitteeService/RequestCommittee
```

### CommitteeRequest

```proto
message CommitteeRequest {
  repeated CommitteeCandidates candidates = 1;
  string channel = 2;
}
```

### CommitteeCandidates

```proto
message CommitteeCandidates {
  string node_id = 1;
  string addr = 2;
  string port = 3;
  bytes  publickey = 4;
  string seed = 5;
  bytes  proof = 6;
}
```

### CommitteeInfo

```proto
message CommitteeInfo {
  string channel_id = 1;
  repeated string member_ids = 2;
  string leader_member_id = 3;
  string timestamp = 4;
}
```

## 데이터 흐름

CEF 기준의 전체 committee election 테스트 흐름은 다음과 같다.

1. blockchain/client node가 CEF의 `JoinNetwork` stream에 접속해 `FinalizedCommittee` 수신 채널을 연다.
2. blockchain/client node가 `RequestCommittee`로 committee candidate 정보를 CEF에 보낸다.
3. CEF는 candidate 정보를 수집하고 MongoDB에 저장한다.
4. candidate가 threshold에 도달하거나 timeout이 발생하면 CEF가 verify node의 `CommitteeService.RequestCommittee`를 호출한다.
5. 이 repository의 `test-verifier`는 CEF로부터 `CommitteeRequest`를 받는다.
6. `test-verifier`는 요청의 `candidates`를 순회하며 `node_id` 목록을 `member_ids`로 만든다.
7. 첫 번째 candidate의 `node_id`를 `leader_member_id`로 지정한다.
8. `CommitteeInfo`를 CEF에 반환한다.
9. 이후 CEF는 선정된 committee member 정보를 기반으로 aggregate public key/commit 생성 및 `FinalizedCommittee` broadcast를 수행한다.

중요한 구현 범위 차이는 다음과 같다.

- 이 repository는 CEF가 수집하는 Schnorr signing commitment나 aggregate commit을 직접 처리하지 않는다.
- 이 repository에는 partial signature 생성, 수집, 검증 RPC나 구현이 없다.
- 이 repository는 `CommitData`, `RequestAggregatedCommit`, `JoinNetwork`, `FinalizedCommittee`를 정의하지 않는다.
- 이 repository의 public surface는 `committee.CommitteeService/RequestCommittee` 하나이다.

## 실행 환경

- Language: Go
- Go module: `test-verifier`
- `go.mod` 기준 Go version: `1.25.0`
- 주요 dependency:
  - `google.golang.org/grpc`
  - `google.golang.org/protobuf`
- gRPC listen port: `50053`
- Docker image name 예시: `verifier:0.1`
- Docker container name: `verify-server1`
- Docker Compose service name: `vs1`
- Docker Compose network name: `bc_interface`

현재 코드에서 별도 environment variable을 읽는 로직은 없다. `Dockerfile`에는 `ARG PORT`가 선언되어 있지만 실제 build/run 로직에서는 사용되지 않는다.

## 실행 방법

기존 README의 로컬 Docker 실행 절차는 유지한다.

```bash
# Docker image 생성
./verifierDockerBuild.sh 0.1

# verify-server1 컨테이너 실행
docker compose up -d
```

`verifierDockerBuild.sh`는 인자로 받은 version을 사용해 다음 형태의 image를 만든다.

```bash
docker build --tag verifier:<version> .
```

`docker-compose.yml`은 기본적으로 `verifier:0.1` image를 사용하므로, compose 파일을 그대로 사용할 경우 다음 명령으로 먼저 image를 빌드해야 한다.

```bash
./verifierDockerBuild.sh 0.1
```

이후 기존 README와 동일하게 [CEF-server](https://github.com/IITP-TestProjects/CEF-server.git)의 README에 따라 CEF를 구동하고, CEF가 이 verify node mock의 `verify-server1:50053` endpoint로 연결되도록 설정한다.

로컬에서 Go binary로 직접 실행하려면 다음과 같이 실행할 수 있다.

```bash
go run .
```

또는 build 후 실행할 수 있다.

```bash
go build -o verifier
./verifier
```

## 설정 정보

### Docker Compose

`docker-compose.yml` 기준 service 설정은 다음과 같다.

- service name: `vs1`
- container name: `verify-server1`
- image: `verifier:0.1`
- port mapping: `50053:50053`
- network: `bc_interface`

CEF-server와 test-verifier를 서로 다른 `docker compose` 프로젝트로 실행하는 경우, 두 컨테이너가 같은 Docker network에 붙어 있어야 한다.  
`bc_interface`가 compose 내부 network로만 생성되면 프로젝트별 prefix가 붙을 수 있으므로, 필요하면 external network로 생성하거나 두 compose 파일의 network 이름을 명시적으로 맞춘다.

### CEF 연동 endpoint

테스트 구성에서 CEF는 다음 endpoint로 이 verify node mock에 접속한다.

```text
verify-server1:50053
```

CEF 실행 옵션이 `-verifynode=verify-server1:50053` 형태라면, 해당 hostname은 Docker DNS에서 `verify-server1` container를 해석할 수 있어야 한다.

## Troubleshooting

### CEF가 verify node에 연결하지 못하는 경우

- `docker compose up -d` 후 `verify-server1` container가 실행 중인지 확인한다.
- CEF container와 `verify-server1` container가 같은 Docker network 또는 서로 라우팅 가능한 network에 있는지 확인한다.
- CEF의 verify node endpoint가 `verify-server1:50053`인지 확인한다.
- host에서 직접 접근하는 경우에는 `localhost:50053` 또는 `127.0.0.1:50053`으로 접근한다.

### `verifier:0.1` image를 찾지 못하는 경우

`docker-compose.yml`은 `verifier:0.1` image를 사용한다. 먼저 다음 명령을 실행한다.

```bash
./verifierDockerBuild.sh 0.1
```

### 빈 candidate 목록 요청

현재 `RequestCommittee` 구현은 첫 번째 candidate의 `NodeId`를 leader로 사용한다. 따라서 빈 candidate 목록이 전달되면 런타임 오류가 발생할 수 있다. CEF에서 verify node를 호출할 때 최소 한 개 이상의 candidate가 포함되어야 한다. CEF에서 verify node를 호출할 때 최소 한 개 이상의 candidate가 포함되어야 한다.

### channel 값이 응답에 반영되지 않는 경우

현재 구현은 요청의 `CommitteeRequest.channel`을 응답의 `channel_id`로 복사하지 않는다. `CommitteeInfo.channel_id`는 항상 `"<test>"`로 반환된다.
