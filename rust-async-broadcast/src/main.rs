
use async_std::io::WriteExt;
use async_std::task;
use async_std::prelude::*;
use async_std::net::{TcpListener, TcpStream};
use async_std::sync::{Arc, Mutex};
use async_std::channel::{Receiver, Sender, self};
use std::collections::HashMap;

//use bytes::{Bytes, BytesMut};
use bytes::{Bytes, BytesMut, Buf, BufMut};
use core::convert::TryInto;

fn main() {
    task::block_on(entrypoint()).expect("failed to initialize!");
}

enum BroadcastCommand {
    AddMember(String, TcpStream),
    DelMember(String),
    SendMessage(TcpStream, Bytes),
    Exit,
}

async fn entrypoint() -> anyhow::Result<()> {
    let listener = TcpListener::bind("0.0.0.0:8080").await?;
    let (sender, recver) = channel::unbounded();
    let broadcaster = task::spawn(broadcaster(recver));
    while let Ok((mut conn, addr)) = listener.accept().await {
        println!("addr: {}", addr);
        // broadcaster한테 등록하라고 명령
        sender.send(BroadcastCommand::AddMember(addr.to_string(), conn.clone())).await?;
        task::spawn(connection(sender.clone(), addr.to_string(), conn.clone()));
    }
    sender.send(BroadcastCommand::Exit).await?;
    broadcaster.await;

    Ok(())
}

async fn broadcaster(recv: Receiver<BroadcastCommand>) {
    // let mut members = Vec::new();
    let mut members = HashMap::new();

    let mut sendCount: i32 = 0;

    // 여기선 event를 받아서 뭐 member에 넣거나 send하거나 등등 하고.
    loop {
        match recv.recv().await.unwrap() {
            BroadcastCommand::AddMember(addr, member) => {
                members.insert(addr, member);
            },
            BroadcastCommand::DelMember(key) => {
                members.remove(&key);
            },
            BroadcastCommand::SendMessage(member, bytes) => {
                println!("members: {}", members.len());
                for (k, v) in &mut members {
                    //if v.as_raw_socket() != member.as_raw_socket() {
                        v.write(&bytes).await;
                        sendCount += 1;
                    //}
                }
                println!("SendMessage: {}", sendCount);
            }
            BroadcastCommand::Exit => {
                break;
            },
        }
    }
}

async fn connection(sender: Sender<BroadcastCommand>, addr: String, mut stream: TcpStream) {
    // connection에서 받아서 broadcaster에 전달하고.
    
    // let b = Bytes::from_static(b"hello");
    // sender.send(BroadcastCommand::SendMessage(stream.clone(), b)).await.unwrap();
    // buf = BytesMut::with_capacity(3);
    let mut buf = BytesMut::new();
    buf.resize(1024, 0);
    let mut recv_buffer2 = BytesMut::with_capacity(50);
    let mut buf = [0; 1024];
    // let headerZZ = 17;
    let headerSize = 16;
    let mut step: u32 = 1;
    let mut need_bytes: usize = headerSize;
    

    loop {
        let n = match stream.read(&mut buf).await {
            // socket closed
            Ok(n) if n == 0 => {
                println!("failed to read from socket;");
                sender.send(BroadcastCommand::DelMember(addr)).await.unwrap();
                break;
            },
            Ok(n) => {
                recv_buffer2.put_slice(&buf[0..n]);
                while recv_buffer2.remaining() >= need_bytes {
                    if step == 1 {
                        let hb = recv_buffer2.copy_to_bytes(need_bytes);
                        let service_code = u32::from_le_bytes(hb[0..4].try_into().unwrap());
                        let length = u32::from_le_bytes(hb[8..12].try_into().unwrap());
                        need_bytes = length as usize;
                        step = 2;
                    } else if step == 2 {
                        println!("[2]step 2: {}", need_bytes);
                        // let hb = recv_buffer2.copy_to_bytes(need_bytes);
                        // if let Err(e) = stream.write_all(&hb[0..need_bytes]).await {
                        //     eprintln!("failed to write to socket; err = {:?}", e);
                        //     return;
                        // }
                        // let b = Bytes::from_static(&hb[0..need_bytes]);
                        let mut sendData = recv_buffer2.clone();
                        sendData.split_off(need_bytes);
                        sender.send(BroadcastCommand::SendMessage(stream.clone(),
                        sendData.freeze())).await.unwrap();
                        need_bytes = headerSize;
                        step = 1;
                    }
                }
            },
            Err(e) => {
                println!("failed to read from socket; err = {:?}", e);
                break;
            }
        };
    }
    // println!("EXIT!!");
}
