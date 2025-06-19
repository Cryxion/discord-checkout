# Discord Bot Checkout
Allow user with role `PaymentManager` to run command
```
!sp <eventid>
```
This will create a button allowing end user/client to click. A DM will be sent to their inbox, for them to further complete their payment.

# Usage
Add your `bot token` at the `.env` with prefix `Bot <token>`


# Permission to enable in discord bot
`Send Messages` = So it can send messages
`View Channels` = So it can see the channel messages


# TODO
- [ ]Stardard RESTFUL return for eventid validation and eventname retrieval  