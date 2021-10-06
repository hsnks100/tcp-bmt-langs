use std::thread;
use std::net::{TcpListener, TcpStream, Shutdown};
use std::io::{Read, Write};
use std::io::BufReader;


use bytebuffer::ByteBuffer;
use core::convert::TryInto;

use bytes::{Bytes, BytesMut, Buf, BufMut};
// extern crate bytebuffer;

fn handle_client(mut stream: TcpStream) {
    let headerSize = 16;
    let mut need_bytes: usize = headerSize;
    let mut step = 1;
    let mut reader = BufReader::new(stream);
    loop {
        // println!("need size: {}", need_bytes);
        let mut buf = vec![0u8; need_bytes];
        if let Ok(size) = reader.read_exact(&mut buf) {
            // println!("{:?}", &buf[0..need_bytes]);
            if step == 1 {
                let service_code = u32::from_le_bytes(buf[0..4].try_into().unwrap());
                let length = u32::from_le_bytes(buf[8..12].try_into().unwrap());
                need_bytes = length as usize;
                step = 2;
                // println!("[0]servicecode: {}, length: {}", service_code, length);

            } else if step == 2 {
                println!("[0]length: {}", need_bytes);
                reader.get_mut().write_all(&buf[0..need_bytes]).unwrap();
                need_bytes = headerSize;
                step = 1;
            }
        } else {
            println!("[0] EOF?");
            break;
        }
    }
    println!("END??");
}

fn main() {
    let listener = TcpListener::bind("0.0.0.0:3335").unwrap();
    println!("Server listening on port 3335");
    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                println!("New connection: {}", stream.peer_addr().unwrap());
                thread::spawn(move|| {
                    handle_client(stream)
                });
            }
            Err(e) => {
                println!("Error: {}", e);
            }
        }
    }
    drop(listener);
}

