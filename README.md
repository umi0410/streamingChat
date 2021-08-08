# Streaming Chat service using GRPC

![assets/preview-1.png](assets/preview-1.png)

## 상상 속의 스트리밍 채팅 서비스

> _(유튜브 라이브 채팅 같은 걸 상상 중...)_
> 
> 실제 채팅 서비스로 운영하기 위함이 아닌 이런 저런 개발적 내용들을 실제로 적용시켜보거나 상상해보기 위한 프로젝트입니다.

* gRPC를 통한 서버 <---> 클라이언트 간의 양방향 통신
* Go 언어 특유의 강력한 동시성 지원을 활용
* Redis stream 및 Redis Pub/Sub을 바탕으로 한 서버간의 통신
* 복잡한 Join이 필요 없고 다량의 데이터를 리얼타임으로 저장해야하므로 No SQL DB 이용
* 채팅방 생성 및 각종 Create, Read 기능은 Java Spring의 MVC와 JPA로 튼튼하고 쉽게 개발
* 트래픽이 몰릴 경우를 대비해 채팅 이벤트는 Redis stream를 버퍼로 두고 워커가 영속화
* k8s + minikube를 이용해 로컬에서도 손쉽게 전체 인프라를 띄울 수 있도록 함.

### 🧐🤔 머릿 속 상상들 ❓⁉️

* gRPC는 어떤 걸까..? RESTful API랑 비교했을 때는 어떨까? 웹소켓과 비교했을 땐 어떨까?
* 메시지 큐(Redis stream)을 이용해서 실제로 마이크로서비스 간의 통신을 해보고싶다.
  * 메시지 큐를 쓰고는 싶은데 Kafka나 RabbitMQ는 메시지 큐를 이용한 마이크로서비스 아키텍쳐를 경험해본다기보단 특정 기술에 대한 지식이 많이 필요한 느낌이고, 관리하기가 쉽지 않다.
  * 관리형인 SNS + SQS 조합을 써보긴 했지만 구성 자체가 한 눈에 안보이는 불편이 있었다. 
  * 실제 운영할 서비스는 아니기도 하고 Redis는 캐싱에도 필수적으로 사용되는 기술이라 관심있었기에 컨테이너로 가볍게 이용하면 될 듯~!
* 얼마나 트래픽이 몰리면 DB가 장애가 날까? 진짜 버퍼가 필요할까?
* Golang이 진짜 동시성 프로그래밍에 있어 성능이 좋을까? 얼마나 좋을까?
* 부하테스트는 어떻게 할 수 있을까?

## gRPC & Protobuf

### 컴파일하는 방법

```text
$ protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pb/chat.proto
```
