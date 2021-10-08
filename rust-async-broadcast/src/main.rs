
use async_std::io::WriteExt;
use async_std::task;
use async_std::prelude::*;
use async_std::net::{TcpListener, TcpStream};
use async_std::sync::{Arc, Mutex};
use async_std::channel::{Receiver, Sender, self};
use std::collections::HashMap;
use futures::join;

// use std::time::Duration;


use std::time::{SystemTime, UNIX_EPOCH};
//use bytes::{Bytes, BytesMut};
use bytes::{Bytes, BytesMut, Buf, BufMut};
use core::convert::TryInto;

enum UsersCommand {
    AddMember(String, Sender<BroadcastCommand>),
    DelMember(String),
    SendAll(Bytes),
}
async fn manage_users(recv: Receiver<UsersCommand>) {
    let mut members: HashMap<String, Sender<BroadcastCommand>>= HashMap::new();
    let mut sendCount: i32 = 0;
    // let mut start = PreciseTime::now();
    let mut start = get_unix_time();
    loop {
        match recv.recv().await.unwrap() {
            UsersCommand::AddMember(addr, member) => {
                members.insert(addr, member);
                // let tt = members.clone();
                // println!("AddMember");
            },
            UsersCommand::DelMember(addr) => {
                // println!("DelMember");
                members.remove(&addr);
            },
            UsersCommand::SendAll(bytes) => {
                for (k, v) in &mut members {
                    let mut sendData = bytes.clone();
                    v.send(BroadcastCommand::SendMessage(sendData)).await.unwrap();
                    sendCount += 1;
                }
                let flowTime = get_unix_time() - start;
                if flowTime >= 1000 {
                    println!("SendMessage: {}/s", (sendCount * 1000) as u128/flowTime);
                    start = get_unix_time();
                    sendCount = 0;
                }
                // sender.send(BroadcastCommand::SendMessage(sendData.freeze())).await.unwrap();
            }

        }
    }
}
fn main() {
    task::block_on(entrypoint()).expect("failed to initialize!");
}

enum BroadcastCommand {
    SendMessage(Bytes),
    Exit,
}

async fn entrypoint() -> anyhow::Result<()> {
    let listener = TcpListener::bind("0.0.0.0:8080").await?;
    let mut members: HashMap<String, Sender<BroadcastCommand>>= HashMap::new();
    let (mgrsender, mgrrecver) = channel::unbounded(); 
    task::spawn(manage_users(mgrrecver));
    while let Ok((mut conn, addr)) = listener.accept().await {
        let (sender, recver) = channel::unbounded();
        let sendChannel = task::spawn(sender_per_client(recver, conn.clone()));
        mgrsender.send(UsersCommand::AddMember(addr.to_string(), sender.clone())).await?;
        task::spawn(connection(addr.to_string(), conn.clone(), sender.clone(), mgrsender.clone()));
    } 
    println!("ENTRY POINT");

    Ok(())
}

fn get_unix_time() -> u128 {
    let start = SystemTime::now();
    let since_the_epoch = start
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards");
    return since_the_epoch.as_millis();
}
async fn sender_per_client(recv: Receiver<BroadcastCommand>, mut v: TcpStream) {
    // let mut members = Vec::new();


    // 여기선 event를 받아서 뭐 member에 넣거나 send하거나 등등 하고.
    loop {
        match recv.recv().await.unwrap() {
            BroadcastCommand::SendMessage(bytes) => {
                // no thread version
                v.write_all(&bytes).await;
            }
            BroadcastCommand::Exit => {
                println!("EXIT MESSAGE");
                break;
            },
        }
    }
}
async fn connection(addr: String, mut stream: TcpStream, sender: Sender<BroadcastCommand>, mgrsender:
                    Sender<UsersCommand>) {
    
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
    
        // sendChannel.await;

    // sender.send(BroadcastCommand::AddMember(addr.to_string(), sender.clone())).await;
    loop {
        let n = match stream.read(&mut buf).await {
            // socket closed
            Ok(n) if n == 0 => {
                println!("failed to read from socket;");
                sender.send(BroadcastCommand::Exit).await.unwrap();
                mgrsender.send(UsersCommand::DelMember(addr)).await.unwrap();
                // sender.send(BroadcastCommand::DelMember(addr)).await.unwrap();
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
                        // println!("[2]step 2: {}", need_bytes);
                        // let hb = recv_buffer2.copy_to_bytes(need_bytes);
                        // if let Err(e) = stream.write_all(&hb[0..need_bytes]).await {
                        //     eprintln!("failed to write to socket; err = {:?}", e);
                        //     return;
                        // }
                        // let b = Bytes::from_static(&hb[0..need_bytes]);
                        // for _ in 1..5 {
                        // }
                        // for (k, v) in members {
                        //     let mut sendData = recv_buffer2.clone();
                        //     sendData.split_off(need_bytes);
                        //     v.send(BroadcastCommand::SendMessage(sendData.freeze())).await.unwrap();
                        // }
                        let mut sendData = recv_buffer2.clone();
                        sendData.split_off(need_bytes);
                        mgrsender.send(UsersCommand::SendAll(sendData.freeze())).await.unwrap();
                        // sender.send(BroadcastCommand::SendMessage(sendData.freeze())).await.unwrap();
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
