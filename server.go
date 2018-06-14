package main

import (
  "github.com/miekg/dns"
  "github.com/go-redis/redis"
  "fmt"
  "log"
  "net"
  "strings"
  "time"
  "os"
)

// TODO
// All Prints must be log
// Read domain list from file

var ClienteGetter = clienteRedis()
var ClienteSetter = clienteRedis()
var Redisserver string = os.Getenv("REDIS_HOST")
var Nameserver string

func clienteRedis() *redis.Client {
  client := redis.NewClient(&redis.Options{
	    Addr: Redisserver + ":6379",
  })

  _, err := client.Ping().Result()
  if err != nil {
    fmt.Println("[ERROR] No se pudo conectar al servidor de Redis!")
    panic(err)
  }

  return client
}

func resolverDominio(dominio string) string {
  m := new(dns.Msg)
  m.Id = dns.Id()
  m.Question = make([]dns.Question, 1)
  m.SetQuestion(dns.Fqdn(dominio), dns.TypeA)
  var respuesta string

  in, _ := dns.Exchange(m, Nameserver + ":53")

  for _, registro := range in.Answer {
    if a, ok := registro.(*dns.A); ok {
      respuesta = a.A.String()
    }
  }

  return respuesta
}

func generarCache(c *redis.Client) {
  for {

    dominios, err := c.Keys("*").Result()
    log.Output(0, "[INFO] Resolviendo direcciones para " + strings.Join(dominios, ",") + ".")
    if err != nil {
      log.Output(0, "[ERROR] Hubo un error al obtener llaves desde Redis!")
    }

    for _, dominio := range dominios {
      err := c.Set(dominio, resolverDominio(dominio), 0).Err()
      if err != nil {
        log.Output(0, "[ERROR] Hubo un error al intentar establecer un par en Redis!")
      }
    }
    time.Sleep(30 * time.Second)
  }

}

func handlerRedis(w dns.ResponseWriter, r *dns.Msg){
  var direccion string

  dominio := strings.TrimSuffix(r.Question[0].Name, ".")
  direccion, err := ClienteGetter.Get(dominio).Result()

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
	srv := &dns.Server{Addr: "0.0.0.0:53", Net: "udp"}
  if err := srv.ListenAndServe(); err != nil {
    panic(err)
  }
}

func main() {

  resolvConfHandler, err := dns.ClientConfigFromFile("/etc/resolv.conf")

  if err != nil {
    log.Fatal(err.Error())
  }

  if len(resolvConfHandler.Servers) > 0 {
    Nameserver = resolvConfHandler.Servers[0]
    log.Output(0, "[INFO] El servidor obtenido es " + Nameserver)
  } else {
    log.Fatal("[ERROR] No hay servidores DNS disponibles configurados en el sistema!")
  }

  go generarCache(ClienteSetter)
  dns.HandleFunc(".", handlerRedis)
  servidor()
}
