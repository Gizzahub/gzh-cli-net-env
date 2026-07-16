# Product Goals (No-PRD)

**Project**: gzh-cli-net-env (library — `NewRootCmd()`, 바이너리 없음)
**Doc Type**: Goals + Constraints + Quality Gates
**Status**: Active
**Last Updated**: 2026-07-16

______________________________________________________________________

## Product Intent

gzh-cli-net-env is a **read-only network environment inspector and profile store**.
It:

- detects the active network (WiFi SSID, IPv4, gateway, DNS, hostname) via
  platform shell-outs,
- scores that state against user-defined YAML profiles to name the current
  environment (home/office/cafe),
- and **describes the network without ever changing it** — 읽기 전용이 정체성이다.

This is a feature-library project — a single PRODUCT.md is sufficient. It
replaces a PRD.

| 제공하는 것 (Is)                              | 되지 않을 것 (Is Not)                       |
| --------------------------------------------- | ------------------------------------------- |
| 네트워크 상태 감지 (SSID·IP·게이트웨이·DNS)   | 네트워크·VPN·DNS·프록시 설정 변경/전환      |
| 가중치 기반 프로필 매칭·YAML 프로필 CRUD      | 데몬·백그라운드 자동 전환                   |
| gzh-cli wrapper가 마운트하는 라이브러리       | 독립 실행 바이너리                          |
| macOS·Linux 감지 백엔드                       | Windows 지원·시크릿 관리                    |

______________________________________________________________________

## Goals (Measurable Targets)

G1. **Detection coverage**

- Target: 2개 플랫폼(macOS·Linux) × WiFi 백엔드 3종(airport, iwgetid→nmcli 폴백)
  — 현재 충족

G2. **Honest status output**

- Target: `status`가 출력하는 컴포넌트는 **전부 실제 감지 결과**여야 한다 (4/4)
- 현재 **1/4** — proxy만 실데이터. WiFi·DNS·VPN은 하드코딩된 상수를 반환하며
  (네트워크가 꺼져 있어도 "connected"), 이를 근거로 Health 점수를 계산한다.
  **이 격차가 본 리포 최우선 과제다.**

G3. **Profile matching determinism**

- Target: 가중치(ssid 100 · gateway 70 · ip_range 50 · hostname 30) + priority로
  단일 승자 결정; 일치 없으면 nil — 현재 충족

G4. **Library-first**

- Target: `main` 패키지 0건, 공개 진입점은 `NewRootCmd()` 하나 (gzh-cli wrapper가
  마운트) — 현재 충족

G5. **Test reliability**

- Target: 커버리지 >= 85% (현재 pkg/netenv 84.5%, cmd/netenv 81.0%, pkg/tui 99.2%)
- 단, `pkg/tui`는 어디서도 import되지 않는 사장 코드다 — 합산 커버리지가 실제
  전달 가치를 과대평가한다. 연결하거나 제거해야 한다.

______________________________________________________________________

## Non-Goals (Explicitly Out of Scope)

- No 네트워크 상태 변경 — WiFi/VPN/DNS/프록시/hosts를 쓰지 않는다 (읽기 전용)
- No Windows 지원 (감지는 명시적으로 unsupported)
- No 독립 실행 바이너리 — 라이브러리로 존재한다 (SOUL 게이트 2)
- No 시크릿 관리 (keychain·vault 연동 없음)
- No 데몬·백그라운드 자동 전환 — `watch`는 포그라운드 폴링 루프다
- No 대역폭·지연 측정

______________________________________________________________________

## Guardrails and Technical Constraints

**Architecture**

- 순수 파싱 함수 + `runtime.GOOS` 플랫폼 dispatch; 미지원 플랫폼은 명시적 에러
- 설정 경로 해석은 `gzh-cli-core/config`에 위임한다 (하드코딩·중복 금지)

**Dependency Boundaries**

- `gzh-cli-core`만 의존 가능; 다른 feature 라이브러리 의존 금지 (GUIDELINES §2)
- 직접 서드파티는 cobra + yaml.v3 뿐 — 추가 시 본 문서에 기록

**Compatibility**

- Go 1.25+ (`go.mod` go 1.25.7; devbox 툴체인 1.26); macOS·Linux

**Safety**

- 모든 `exec.Command`는 읽기 전용 조회다 (airport/iwgetid/nmcli/route/ip route)
- sudo·권한 상승을 사용하지 않는다 (0건)
- 프로필 파일은 0600, 디렉터리는 0750; 프로필명 검증이 경로 탈출을 차단한다
- **미비**: `profile delete`는 확인 프롬프트·백업 없이 즉시 삭제한다;
  `ProxyAuth`의 자격증명은 평문 YAML로 저장된다 (`${PROXY_USER}` 확장 코드 없음)

**Baseline**

- GUIDELINES §3 베이스라인 충족 — `Makefile`·`.golangci.yml`(v2)·CI·`LICENSE`(MIT,
  소스의 SPDX 헤더와 일치)·문서·본 PRODUCT.md 보유

______________________________________________________________________

## Quality Gates (Release Readiness)

**Build and Lint**

- `make check` (fmt + lint + test) pass with no warnings

**Testing**

- `go test ./... -cover` pass; 커버리지 >= 85%

**Correctness**

- `status`의 모든 컴포넌트가 실제 감지 결과를 반영한다 (G2 — 현재 미충족)

**Docs**

- README·컨텍스트 문서의 명령·플래그가 실제와 일치한다 (현재 미충족: 바이너리
  존재 주장, `status --json`은 실제로 `--format json`)

______________________________________________________________________

## Decision Rules

- **하드코딩된 상태값을 사실인 것처럼 반환하는 코드는 머지될 수 없다** — 감지
  도구의 신뢰가 유일한 자산이다
- 네트워크 상태를 **변경**하는 기능은 읽기 전용 정체성을 깨므로 오너 승인을 요구한다
- 타입만 선언하고 동작이 없는 스키마 확장은 추가하지 않는다 (Docker·Kubernetes·
  Bandwidth 등 기존 미사용 타입이 이미 부담이다)
- 새 기능은 SOUL.md 4-게이트(틈 · 라이브러리 · 대량/전환 · 날카로움)를 통과해야 한다
- Quality Gates 미충족 시 릴리스는 차단된다

______________________________________________________________________

**End of Document**
