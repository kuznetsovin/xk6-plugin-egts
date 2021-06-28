# xk6-plugin-egts

It k6 extention for stress testing EGTS protocol receivers.

## Build

To build a `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- [Git](https://git-scm.com/)

Then, install [xk6](https://github.com/k6io/xk6) and build your custom k6 binary with the Kafka extension:

1. Install `xk6`:
  ```shell
  $ go install github.com/k6io/xk6/cmd/xk6@latest
  ```

2. Build the binary:
  ```shell
  $ xk6 build --with github.com/kuznetsovin/xk6-plugin-egts@latest
  ```

## Example

```javascript
import egts from "k6/x/egts";

export let options = {
    vus: 2,
    iterations: 2,
};

// client testing tracks, where key is VU number
// point of track is array [lat, lon, sens_value, fuel_level]
// if sens_value or fuel_level equals 0 then sending simple packet whith coordinate section only 
const data = {
    1: [[55.55389399769574, 37.43236696287812, 1000, 1000], [55.55389399769574, 37.43236696287812, 1000, 1000]],
    2: [[55.55389399769574, 37.43236696287812, 1000, 1000], [55.55389399769574, 37.43236696287812, 200, 200]]
}

//for each VU open connection for emulating device
export default () => {
    let client = egts.newClient("127.0.0.1:6000", __VU);
    data[__VU].forEach((rec) => {
        egts.sendPacket(client, ...rec)
    })

    egts.closeConnection(client)
};
```

Result output:

```bash

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: example/example.js
     output: -

  scenarios: (100.00%) 1 scenario, 2 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 2 iterations shared among 2 VUs (maxDuration: 10m0s, gracefulStop: 30s)


running (00m00.0s), 0/2 VUs, 2 complete and 0 interrupted iterations
default ✓ [======================================] 2 VUs  00m00.0s/10m0s  2/2 shared iters

     data_received...............: 116 B 16 kB/s
     data_sent...................: 352 B 49 kB/s
     egts_packets................: 4     552.638851/s
     egts_packets_process_time...: avg=2.41µs min=565ns  med=2.32µs max=4.42µs p(90)=3.94µs p(95)=4.18µs
     iteration_duration..........: avg=5.47ms min=5.35ms med=5.47ms max=5.59ms p(90)=5.57ms p(95)=5.58ms
     iterations..................: 2     276.319425/s
```