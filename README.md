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
    scenarios: {
        scenario_1: {
            executor: 'shared-iterations',            
            vus: 2,
            iterations: 4,
        },
        scenario_2: {
            executor: 'per-vu-iterations',            
            vus: 2,
            iterations: 2,
        },
    }
};

// client testing tracks, where key is VU number
// point of track is array [lat, lon, sens_value, fuel_level]
// if sens_value or fuel_level equals 0 then sending simple packet whith coordinate section only 
const data = {
    0: [[55.55389399769574, 37.43236696287812, 1000, 1000], [55.55389399769574, 37.43236696287812, 1000, 1000]],
    1: [[55.55389399769574, 37.43236696287812, 1000, 1000], [55.55389399769574, 37.43236696287812, 200, 200]]
}

//for each VU open connection for emulating device
export default () => {
    let client = egts.newClient("127.0.0.1:6000", __VU);
    data[__VU%2].forEach((rec) => {
        egts.sendPacket(client, ...rec)
    })

    client.close()
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

  scenarios: (100.00%) 2 scenarios, 4 max VUs, 10m30s max duration (incl. graceful stop):
           * scenario_1: 4 iterations shared among 2 VUs (maxDuration: 10m0s, gracefulStop: 30s)
           * scenario_2: 2 iterations for each of 2 VUs (maxDuration: 10m0s, gracefulStop: 30s)


running (00m00.2s), 0/4 VUs, 8 complete and 0 interrupted iterations
scenario_1 ✓ [======================================] 2 VUs  00m00.2s/10m0s  4/4 shared iters
scenario_2 ✓ [======================================] 2 VUs  00m00.2s/10m0s  4/4 iters, 2 per VU

     data_received...............: 464 B  2.0 kB/s
     data_sent...................: 1.4 kB 6.2 kB/s
     egts_packets................: 16     69.993701/s
     egts_packets_process_time...: avg=56.47µs  min=1.65µs med=20.38µs  max=200.24µs p(90)=185.85µs p(95)=193.26µs
     iteration_duration..........: avg=113.44ms min=3.92ms med=113.77ms max=222.6ms  p(90)=222.01ms p(95)=222.31ms
     iterations..................: 8      34.99685/s
```