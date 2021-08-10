# Streaming Chat service using GRPC

![k9s](https://user-images.githubusercontent.com/33250725/128679725-27e3d8d8-601b-4b6a-8e2c-081531e29fff.gif)

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

* gRPC는 어떤 걸까..? RESTful API나 GrpahQL이랑 비교했을 때는 어떨까? 웹소켓과 비교했을 땐 어떨까?
* 메시지 큐(Redis stream)을 이용해서 실제로 마이크로서비스 간의 통신을 해보고싶다.
  * 메시지 큐를 쓰고는 싶은데 Kafka나 RabbitMQ는 메시지 큐를 이용한 마이크로서비스 아키텍쳐를 경험해본다기보단 특정 기술에 대한 지식이 많이 필요한 느낌이고, 관리하기가 쉽지 않다.
  * 관리형인 SNS + SQS 조합을 써보긴 했지만 구성 자체가 한 눈에 안보이는 불편이 있었다. 
  * 실제 운영할 서비스는 아니기도 하고 Redis는 캐싱에도 필수적으로 사용되는 기술이라 관심있었기에 컨테이너로 가볍게 이용하면 될 듯~!
* 얼마나 트래픽이 몰리면 DB가 장애가 날까? 진짜 버퍼가 필요할까?
* Golang이 진짜 동시성 프로그래밍에 있어 성능이 좋을까? 얼마나 좋을까?
* 양방향 통신을 하는 경우 실제 운영할 때에는 어떻게 Gracefully shutdown을 할 수 있을까? Golang에서는 Context를 통해 최대한 Graceful하게 할 수 있을 것 같은데..
* 부하테스트는 어떻게 할 수 있을까?

## gRPC & Protobuf

### 컴파일하는 방법

```text
$ protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pb/chat.proto
```

### gRPC의 Stream을 이용한 채팅 이벤트 스트리밍

## Redis

### Redis Pub/Sub을 이용한 데이터 전송

* 각 서버는 자신과 연결되어 있는 클라이언트 정보를 갖고 있다. 해당 클라이언트들에겐 자신(서버)이 새로운 메시지 이벤트를 전달해주어야한다.
* 어떤 채팅 서버 replica에서 채팅 이벤트가 발생한 경우 또 다른 채팅 서버 replica는 그 이벤트를 어떻게 알고 자신과 연결된 클라이언트들에게 전달할 수 있을까?
* 이때 Redis Pub/Sub을 사용하면 좋다!
* Redis stream을 쓰는 건 어떨까?
  * Consumer group 관리나 ACKnowledge 등의 기능을 수행하는 Redis stream보다는 Pub/Sub이 Real-time 성격의 이 스트리밍형 채팅에 더 어울릴 것 같음.
  * 만약 Pub/Sub을 이용했다가 서버가 죽어서 클라이언트에게 알림을 보내줄 수 없다면?
    * 일단은 클라이언트는 자신과 연결된 서버가 죽으면 재빨리 다른 서버와 커넥션을 맺도록 되어있음.
    * 그 사이에 발생한 채팅은? - 사실 그냥 무시한다. 이런 것까지 다 신경을 쓰는 건 다양한 기술을 적용한 아키텍쳐를 상상해보고 적용시켜보는 이 프로젝트와는 점점 거리가 멀어지고
      채팅 도메인을 개발하면서 도메인 로직이나 비즈니스 로직에 점점 집중하게 될 것 같기 때문. 최대한 리얼타임 스트리밍형이라는 것에 초점을 둔다!
    
### Redis stream을 이용한 데이터 전송

* Real-time 성격이 아닌 단순히 채팅 데이터를 DB에 영속화 시키는 경우에는 재빨리 채팅 이벤트를 클라이언트들에게 뿌려주는 것보다
  해당 이벤트가 잘 처리됐는지 ACKnowledge할 수 있어야 함. 따라서 Pub/Sub 보다는 Redis의 Stream을 이용해 Kafka처럼 사용하는 것이 좋다.
* Pub/Sub의 경우 Consumer Group이라는 개념이 필요 없이 각 백엔드 replica들이 동일한 이벤트를 모두 전달받아야했지만
  채팅 데이터를 DB에 영속화 시키는 작업은 Consumer group의 개념을 이용해 group 내의 replica 중 한 명만 이벤트를 올바르게 처리해주면 된다.
  따라서 Consumer group을 이용하기 위해 Pub/Sub이 아닌 Redis stream을 이용한다.

### Redis의 TTL을 이용한 실시간 인기 단어

> 사실 별로 필요한 기능은 아닌데 ㅎㅎ.. 다양한 Redis의 유즈케이스가 궁금해서 상상해봤다.

보통 Realtime leaderboard라는 컨셉으로 Sorted Set을 이용해 리얼타임 랭킹 서비스를 소개하는 것 같은데
이 경우는 전체 기간 동안에 대한 Ranking만 제공하고 특정 기간 동안의 랭킹은 제공할 수 없다. Sorted Set에 TTL이 없기 때문이기도 하고,
결국 특정 기간 동안의 이벤트로만 랭킹을 정렬하려면 이용되는 각 이벤트의 정보가 있어야한다. 근데 마침 이에 대해 나의 상상과 비슷한 해결책을 다룬 [스택오버 플로우 글](https://stackoverflow.com/a/55584571/9471220)이 있었다.

* Redis의 Expiration time(TTL)을 이용하면 인기 단어 집계에 사용되는 단어 이벤트 데이터가 손쉽게 알아서 특정 시간 이후 삭제되도록 할 수 있다.
* 그럼 특정 기간(예를 들어 지난 1분) 간의 인기 단어를 보기 위해선 TTL을 1분으로 설정해두고 남아있는 단어 이벤트만으로 집계하면 된다.
* 단점은 Sorted Set을 이용할 때는 랭킹 순위 자체가 데이터를 저장할 때 저장된다는 것이었는데, 이 방법을 이용하면 정렬은 매 쿼리 때마다 애플리케이션에서 담당해야한다.  
* `hotWord@{{word}}@{{username}}: 0` 이런 식으로 Key를 설정하고 empty value를 담은 뒤 `hotWord@{{word}}@*`에 해당하는 key의 개수를 세어 정렬하면 된다.
* _"Keeping Redis simple"_ 이라는 철학 때문인지 다소 `@` 로 'howWord 랭킹 산출에 사용되는 키이며 `{{word}}`라는 단어에 대한 `{{username` 유저의 채팅 기록이다'
  라는 의미를 나타내는 것이 다소 깔끔하진 않아보일 수 있지만 뭐 RDB나 다른 DB를 썼으면 TTL을 위해 또 다른 이런 저런 고생을 했어야할 것 같긴 하니까 Redis를 용서해주는 걸로 하겠다.

## 참고 자료

샘플 채팅 데이터 - https://github.com/songys/Chatbot_data 