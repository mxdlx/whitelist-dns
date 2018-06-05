package main

import (
  "github.com/miekg/dns"
  "github.com/go-redis/redis"
  "fmt"
  "net"
  "strings"
  "time"
)

// TODO
// All Prints must be log
// Read domain list from file

var clienteGetter = clienteRedis()
var clienteSetter = clienteRedis()

func clienteRedis() *redis.Client {
  client := redis.NewClient(&redis.Options{
	    Addr: "localhost:6379",
	    Password: "",
	    DB: 0,
  })

  pong, err := client.Ping().Result()
  if err != nil {
    fmt.Println("[ERROR] No se pudo conectar al servidor de Redis.")
    panic(err)
  }
  fmt.Println(pong)
  return client
}

func resolverDominio(dominio string) string {
  m := new(dns.Msg)
  m.Id = dns.Id()
  m.Question = make([]dns.Question, 1)
  m.SetQuestion(dns.Fqdn(dominio), dns.TypeA)
  respuesta := ""

  in, _ := dns.Exchange(m, "8.8.8.8:53")

  for _, registro := range in.Answer {
    if a, ok := registro.(*dns.A); ok {
      respuesta = a.A.String()
    }
  }

  return respuesta
}

func generarCache(c *redis.Client) {
  for {
    var dominios []string
    dominios = append(dominios, "www.google.com", "www.amazon.com")

    for _, dominio := range dominios {
      err := c.Set(dominio, resolverDominio(dominio), 0).Err()
      if err != nil {
        panic(err)
      }
    }
    time.Sleep(5 * time.Second)
  }

}

func handleDnsRequests(w dns.ResponseWriter, r *dns.Msg){
  var direccion string

  dominio := strings.TrimSuffix(r.Question[0].Name, ".")
  direccion, err := clienteSetter.Get(dominio).Result()

  if err != nil {
    fmt.Printf("[ALERTA] Dominio %s no permitido - Redis dice: %s\n", dominio, err)
    direccion = "127.0.0.1"
  }

  m := new(dns.Msg)
  m.SetReply(r)
  rr := &dns.A{
    Hdr: dns.RR_Header{Name: dns.Fqdn(dominio), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
    A: net.ParseIP(direccion),
  }
  m.Answer = append(m.Answer, rr)
  w.WriteMsg(m)
}

func servidor() {
  srv := &dns.Server{Addr: "127.0.0.1:2053", Net: "udp"}
  if err := srv.ListenAndServe(); err != nil {
    panic(err)
  }
}

func main() {
  go generarCache(clienteSetter)
  dns.HandleFunc(".", handleDnsRequests)
  servidor()
}
