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
sudo ./packetloss -p 29501,29502,29503,29504,29505.29506,29507,29508 -r 0.25
sudo ./packetloss -f

## TODO
- Update requests at elevators
    - Send new requests to elevator (tcp/udp?)
    - Fsm_OnRequestButtonPress (maybe rename this? I think it should be called when new orders are recieved?)
    - Maybe modify this function to also update requests?
    - Right now it only takes new buttonpresses, and does not remove wrong ones
    - Bit mask?
    - Maybe use bit mask first, then run the function without button
    - Cost function does return the floor you are at
- Cab calls should work even offline
- Packet loss should only cause delay in all lights being synchronized
- Kill oldest primary if two is present
    - Store time of beoming primary
    - Send it with primary alive
    - Only needs to check for time in primary
    - Should current queue be transfered? I think no?
- Must kill backup when connection to it is lost

NEED TO FIX STUFF WITH BACKUP DYING AND COMING BACK ALIVE
https://prod.liveshare.vsengsaas.visualstudio.com/join?87F7C257C8255F2FDA493D58D88A9BD5476F