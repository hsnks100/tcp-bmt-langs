use tokio::net::TcpListener;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use bytebuffer::ByteBuffer;

use core::convert::TryInto;
use bytes::{Bytes, BytesMut, Buf, BufMut};
#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let listener = TcpListener::bind("0.0.0.0:3333").await?;

    loop {
        let (mut socket, _) = listener.accept().await?;

        tokio::spawn(async move {
            let mut buf = [0; 1024];
            let headerSize = 16;
            let mut need_bytes: usize = headerSize;
            let mut step = 1;
            let mut recv_buffer2 = BytesMut::with_capacity(50);
            loop {
                let n = match socket.read(&mut buf).await {
                    // socket closed
                    Ok(n) if n == 0 => return,
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
                                let hb = recv_buffer2.copy_to_bytes(need_bytes);
                                if let Err(e) = socket.write_all(&hb[0..need_bytes]).await {
                                    eprintln!("failed to write to socket; err = {:?}", e);
                                    return;
                                }
                                need_bytes = headerSize;
                                step = 1;
                            }
                        }
                        // println!("END??");
                    },
                    Err(e) => {
                        eprintln!("failed to read from socket; err = {:?}", e);
                        return;
                    }
                };

                // Write the data back
                // if let Err(e) = socket.write_all(&buf[0..n]).await {
                //     eprintln!("failed to write to socket; err = {:?}", e);
                //     return;
                // }
            }
        });
    }
}

