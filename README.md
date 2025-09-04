# Verify Node (인천대 파트 테스트용 모사 컨테이너)

제안서 상 Blockchain - CEF Server - Verify Node 전체 아키텍처 상태에서 테스트를 위한 verify node 모사 컨테이너 생성을 위한 프로젝트이다. 

---

### 구동방법(로컬 도커에서 컨테이너 간 연결)

```bash
#도커 이미지를 생성
~$ ./verifierDockerBuild.sh 0.1

~$ docker compose up -d
```

이후, [CEF-server](https://github.com/IITP-TestProjects/CEF-server.git) 에서 README를 따라 구동하고 본 컨테이너와 연결시도를 하면 연결된다.