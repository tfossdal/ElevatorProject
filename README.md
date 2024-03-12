# Elevator Project for TTK4145



## Port list (starts with 295##)
- 01 - Primary alive
- 02 - Backup alive
- 03 - Dumb/Primary communication
- 04 - Distribute orders
- 05 - Light light
- 06 - Primary/Backup communication
- 07 - Primary/Primary cab orders transmit
- 08 - Primary/Primary cab orders retransmit


## Order Format (UDP)
- n,a,b
    - n - new order/ID
    - a is floor, b is button type

## I'm Alive Format (UDP)
- s,s,d,f
    - state of elevator
    - state (0=idle, 1=moving, 2=doorOpen)
    - direction (1=up, -1=down, 0=stop)
    - floor (0 - m-1)

## OrderToBackup (TCP)
- n,ID,f,b,
    - new order
    - ID (num as string)
    - floor (0 - m-1)
    - button (0=up, 1=down, 2=cab)
## Packetloss command
sudo ./packetloss -p 29501,29502,29503,29504,29505,29506,29507,29508 -r 0.25
sudo ./packetloss -f
