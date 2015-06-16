package main

import (
    "os"
    "io"
    "fmt"
    "log"
    "flag"
    "runtime"
    "github.com/cheggaaa/pb"
    "github.com/dustin/go-humanize"
    "github.com/qedus/osmpbf"
)

func main() {
    ncpu := flag.Int("ncpu", 1, "number of CPU")
    flag.Parse()
    runtime.GOMAXPROCS(*ncpu)
    for _, file := range flag.Args() {
        worker(file)
    }
}

func worker(file string) { 
    f, err := os.Open(file)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    stat, _ := f.Stat()
    filesiz := int(stat.Size()/1024)

    d := osmpbf.NewDecoder(f)
    err = d.Start(runtime.GOMAXPROCS(-1))
    if err != nil {
        log.Fatal(err)
    }

    var nc, wc, rc, i int64
    progressbar := pb.New(filesiz).SetUnits(pb.U_NO)
    progressbar.Start()
    for i = 0; ; i++ {
        if v, err := d.Decode(); err == io.EOF {
            break
        } else if err != nil {
            log.Fatal(err)
        } else {
            switch v := v.(type) {
            case *osmpbf.Node:
                nc++
            case *osmpbf.Way:
                wc++
            case *osmpbf.Relation:
                rc++
            default:
                log.Fatalf("unknown type %T\n", v)
            }
        }
        if i % 131072 == 0 {
            progressbar.Set(int(d.GetTotalReadSize()/1024))
        }
    }
    progressbar.Set(filesiz)
    progressbar.Finish()
    fmt.Printf("Nodes: %s, Ways: %s, Relations: %s\n", humanize.Comma(nc), humanize.Comma(wc), humanize.Comma(rc))
}
