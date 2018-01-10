package main

import (
        "context"
        "fmt"

        "bytes"
        "io"
        "log"
        "net"
        "net/http"
        "strconv"
        "strings"
        "time"

        "github.com/docker/docker/client"
)

type NodeInfo struct {
        Name       string
        SwarmInfo  *swarmInfo
}

type swarmInfo struct {
        NodeId     string
        NodeAddr   string
}

func getSwarmInfo(cli *client.Client) *swarmInfo {
        info, err := cli.Info(context.Background())
        if err != nil {
          panic(err)
        }
        return &swarmInfo{
                NodeId:   info.Swarm.NodeID,
                NodeAddr: info.Swarm.NodeAddr,
        }
}

func getNodeInfo(cli *client.Client) *NodeInfo {
        info, err := cli.Info(context.Background())
        if err != nil {
          panic(err)
        }
        return &NodeInfo{
                Name:      info.Name,
                SwarmInfo: getSwarmInfo(cli),
        }

}

//
func readLineWith(buf *[]byte, delim string, match string)(res string) {
        if len(*buf) < 0 {
                return ""
        }
        lines := bytes.Split(*buf, []byte(delim))
        for i, v := range lines {
                if strings.HasPrefix(string(v), match) {
                        fmt.Println("The regex(" + match + ") line is on line " + string(i))
                        return string(v)
                }
        }
        return ""
}

func parseGpuInfo(info string)(node string, task string, index string) {
        ss := strings.Split(info, ",")
        if len(ss) != 3 {
                return "", "", "-1"
        }
        node = ss[0]
        task = ss[1]
        ret_idx, err := strconv.Atoi(ss[2])
        if err != nil {
                fmt.Println("Cant format " + ss[2] + "to a Int value...")
                return node, task, "-1"
        }
        index = strconv.Itoa(ret_idx)
        return node, task, index
}

func httpRequestGPU(uri string, node string, task string)(ret_node string, ret_task string, ret_index string) {
        var netTransport = &http.Transport{
            Dial: (&net.Dialer{
            Timeout: 5 * time.Second,
            }).Dial,
            TLSHandshakeTimeout: 5 * time.Second,
            DisableKeepAlives: true,
        }
        var netClient = &http.Client{
            Timeout: time.Second * 10,
            Transport: netTransport,
        }

        // Serve 256 bytes every second.
        req, err := http.NewRequest("GET", uri + "?node=" + node + "&task=" + task, nil)
        if err != nil {
                log.Fatal(err)
        }

        resp, err := netClient.Do(req)
        if err != nil {
                log.Fatal(err)
        }
        defer resp.Body.Close()

        for {
                content := bytes.NewBuffer([]byte(""))
                body_size, err := io.CopyN(content, resp.Body, 1024)
                if err == io.EOF {
                        log.Printf("Received %d msg + %s", body_size, string(content.Bytes()))
                        body := content.Bytes()
                        info := readLineWith(&body, "\n", "GPU-info")

                        return parseGpuInfo(info)
                } else if err != nil {
                        log.Fatal(err)
                }
        }

        return node, task, "-1"
}

func requestResource(server string, name string)(ret_node string, ret_task string, ret_idx string) {
        //runtime.Breakpoint()
        cli, err := client.NewEnvClient()
        if err != nil {
                panic(err)
        }
        nodeInfo := getNodeInfo(cli)
        node := nodeInfo.SwarmInfo.NodeAddr
        task := name

        ret_node, ret_task, ret_idx = httpRequestGPU("http://" + server, node, task)

        log.Println("Client has fetched GPU index >> " + ret_node, ":", ret_task, ":", ret_idx)
        return ret_node, ret_task, ret_idx
}

func checkServerAddress(server string)(res bool) {
       sp := strings.Split(server, ":")
       if len(sp) != 2 {
             return false
       }
       return true
}
