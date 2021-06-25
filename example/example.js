import egts from "k6/x/egts";

export let options = {
    vus: 2,
    iterations: 2,
};

// testing client tracks, where key is VU number
const data = {
    1: [[55.55389399769574, 37.43236696287812, 1000, 1000], [55.55389399769574, 37.43236696287812, 1000, 1000]],
    2: [[55.55389399769574, 37.43236696287812, 1000, 1000], [55.55389399769574, 37.43236696287812, 200, 200]]
}

//for each VU open connection
export default () => {
    let client = egts.newClient("127.0.0.1:6000", __VU);
    data[__VU].forEach((rec) => {
        egts.sendPacket(client, ...rec)
    })

    egts.closeConnection(client)
};

