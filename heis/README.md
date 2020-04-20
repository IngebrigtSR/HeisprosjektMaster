
# TTK4145 Heisprosjekt

Vårt system baserer seg på et peer to peer nettverk hvor alle heisene på nettverket har informasjon om alle de andre til en hver tid i form av en lokalt lagret logg over alle heisene. Heisene fordeler ordre mellom seg ved hjelp av en cost funksjon og broadcaster sin nye oppdaterte logg til alle de andre som så plukker ned denne og oppdaterer sine lokale logger slik at alle så fort som mulig får den mest oppdaterte informasjonen.

## Moduler

### Logmanager

Modulen håndterer den lokalt lagrede loggen som skal sendes og mottas til og fra andre heiser. Loggen er en tabell der hvert element er en struct med all nødvendig informasjon for hver unike heis. Dermed kan hver enkelt heis fungere på egen hånd i henhold til kravsspesifikasjonen basert på informasjonen i loggen.

### Orderhandler

Inneholder alt av funksjonalitet for å håndtere fordeling av ordre til de forskjellige heisene samt funksjoner for å slette utførte ordre fra loggen og også godta ordre som er blitt tildelt fra andre heiser. Fordeling av ordrene baserer seg på en cost funksjon som rett og slett teller hvor mange etasjer heisen må passere eller stoppe i før den vil nå en gitt ordre, basert på denne kostnaden vil da den "billigste" heisen på system bli tildelt ordren.


### FSM

Håndterer logikken for å kunne kjøre heisen og reagere på hendelser på hardware. Det er her programmet reagerer på knappetrykk og etasjesensorer, samt oppdatere og sette diverse knappelys og motorretning på heisen avhengig av ordre det har liggende i loggen. 

### Config

Config inneholder diverse konstanter som brukes av de andre moduler. Her er for eksempel totalt antall heiser og etasjer på systemet bestemt og porter for kommunikasjon over nettverket er også satt her.

## Utdelt kode

### Elevio

Dette er driveren som kommuniserer med hardware på heisen. Her settes fysisk alt av lys og motorretning og detekterer knappetrykk og etasjesenorer.

### Network

Inneholder funksjoner for å sende og motta meldinger over nettet. Her ligger også peers som oppdager hver gang noen nye heiser dukker opp på nettet. Mer spesifikt om nettverksmodulen kan leses her: https://github.com/TTK4145/Network-go
