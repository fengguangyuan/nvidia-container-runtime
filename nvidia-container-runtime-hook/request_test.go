package maintest

import (
        "testing"
)

func TestDefaultNodeRequest(t *testing.T) {
        node, task, index := requestResource("10.212.51.53:21497", "sadoioasihdg")
        if index == "-1" {
                t.Errorf("Recived invalid GPU info : %s : %s : %d", node, task, index)
        }
        t.Log("Recived GPU info : %s : %s : %d", node, task, index)
}

func TestSpecifiedNodeRequest(t *testing.T) {
        node, task, index := httpRequestGPU("http://127.0.0.1:21497", "127.0.0.1", "sdionewingi")
        if index == "-1" {
                t.Errorf("Recived invalid GPU info : %s : %s : %d", node, task, index)
        }
        t.Log("Recived GPU info : %s : %s : %d", node, task, index)
}

func TestInvalidNodeRequest(t *testing.T) {
        node, task, index := httpRequestGPU("http://127.0.0.1:21497", "127.0.0.1", "sdionewingi")
        if index != "-1" {
                t.Errorf("Recived valid GPU info : %s : %s : %d", node, task, index)
        }
        t.Log("Recived GPU info : %s : %s : %d", node, task, index)
}

func TestValidNodeRequest(t *testing.T) {
        node, task, index := httpRequestGPU("http://10.212.51.53:21497", "127.0.0.2", "sdionewingi")
        if index != "-1" {
                t.Errorf("Recived valid GPU info : %s : %s : %d", node, task, index)
        }
        t.Log("Recived GPU info : %s : %s : %d", node, task, index)
}
