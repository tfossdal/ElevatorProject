# Elevator Project for TTK4145



## Port list (starts with 295##)
- 01 - Primary alive
- 02 - Backup alive
- 03 - Dumb elevator alive
- 05 - Light light
- 06 - Primary/Backup communication

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

## TODO
- Must kill backup when connection to it is lost
https://prod.liveshare.vsengsaas.visualstudio.com/join?236EB79E1FD3C17F58BA537C058F1AB346CA