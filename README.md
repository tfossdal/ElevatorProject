# Elevator Project for TTK4145



## Port list (starts with 295##)
- 01 - Primary alive
- 02 - Backup alive
- 03 - Dumb elevator alive
- xx 05 - Add order
- 06 - Primary/Backup communication

## Order Format (UDP)
- n,a,b
    - no - new order
    - a is floor, b is button type

- Must kill backup when connection to it is lost
