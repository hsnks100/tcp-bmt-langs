# tcp-bmt-langs

# test 방법 

server:

```
cd rust-async-broadcast
cargo build --release && ./target/release/kk
```

client:

```
cd attacker
go build && ./kk 8080
```

서버측에서 초당 몇회 쏘는지 나옴.
