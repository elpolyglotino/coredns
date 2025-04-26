package hosthealth

import (
    "fmt"
    "net"
    "sync"
    "time"
)

/**
 * @author Hossein Boka <i@Ho3e.in>
 */

func (hr *HealthRecords) runHealthChecks() {
    for {
        hr.Mutex.Lock()
        for name, recs := range hr.Records {
            for _, rec := range recs {
                alive := checkIP(rec.IP)
                if rec.Alive != alive {
                    fmt.Printf("[hosthealth] %s -> %s: %v\n", name, rec.IP, alive)
                }
                rec.Alive = alive
            }
        }
        hr.Mutex.Unlock()
        time.Sleep(5 * time.Second)
    }
}

func checkIP(ip string) bool {
    conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, 80), 2*time.Second)
    if err != nil {
        return false
    }
    conn.Close()
    return true
}
