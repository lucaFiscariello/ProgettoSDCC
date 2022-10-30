# ProgettoSDCC

## Descrizione

In questa applicazione un certo numero di nodi concorre a popolare una pagina web con notizie di attualità. Ogni nodo scaricherà notizie relative a un tema particolare
e cercherà di pubblicare queste informazioni dinamicamente. In particolare ogni nodo vuole pubblicare una notizia per volta sulla pagina web ma tale scrittura non può avvenire in maniera concorrente. 
Per questo motivo si rende necessario l'utilizzo di un algoritmo per la mutua esclusione che garantisce che un solo nodo per volta possa scrivere la propria notizia senza creare conflitti.


## Esecuzione Algoritmo Centralizzato o Lamport distribuito

```bash
cd Algoritmi/LamportDistribuito
docker compose up --build
```

```bash
cd Algoritmi/AutorizzazioneCentralizzata
docker compose up --build
```

## Esecuzione Token distribuito
L'algoritmo di token distributo è configurato per essere deployato su un istanza ec2. Con le configurazioni attuali non è possibile lanciarlo in locale. Il comando per lanciare l'algoritmo su un'istanza di ec2 è il seguente.

NB: è necessario fornire la propria chiave privata.

```bash
cd Ansible
ansible-playbook -v -i host.ini deploy.yaml
```


## Risultati

Una volta lanciato il comando è possibile visitare _http://localhost:8080/_ e osservare come dinamicamente il sito venga popolato da notizie e continuamente aggiornato.
Un esempio di come la pagina web apparità dopo alcuni cicli di esecuzione è il seguente.

![](https://github.com/lucaFiscariello/ProgettoSDCC/blob/d50dfb2d1db19d22de57083bbabfbcc742b5fd7e/Token%20distribuito/WebServer/webSite/template.png)
