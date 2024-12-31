# lbx-telebot
bot for our group community

*Available commands:*

`/help` or `/h`

Show this help.

`/ver` or `/v`

Show the current version.

`/profile` _NAME_ _PERIOD_

Show the stored profile of the requester or another user.

Options:
- _NAME_ - target chat participant
- _PERIOD_ - custom period of statistic (e.g. 7d, 72h), should be > 0

`/top` _NUM_ _PERIOD_

Show top users.

Options:
- _NUM_ - custom number of positions to show, should be > 0
- _PERIOD_ - custom period of statistic (e.g. 7d, 72h), should be > 0

`/bottom` _NUM_ _PERIOD_

Show reversed rating

Options:
- _NUM_ - custom number of positions to show, should be > 0
- _PERIOD_ - custom period of statistic (e.g. 7d, 72h), should be > 0

`/topic` _text_

Set new title in the group

`/event`

Command for event. Send command without params for detailed instructions.

`/today`

Returns what happened on this day

`/newyear` _TIMEZONE_
Returns time left until the new year
Options:
_TIMEZONE_ - custom timezone. E.g `Europe/Podgorica`

---

`/event` create _NAME_

Create new event with _NAME_ option. It could be sent in group chat or in a direct chat with Valera.
You should have admin rights.

Option is required:
- NAME - Uniq name for new event. Should be one word with chars and digits only

`/event` list \[_-a_ | _all_]

Show all active event. It could be sent in group chat or in a direct chat with Valera.

Options:
- _-a_ (or "_all_") shows all events either open or finished

`/event` info _NAME_

Show the event information and bets

Option is required:
- NAME - Uniq name for new event. Should be one word with chars and digits only

`/event` my _NAME_

Show your personal bet in the particular event

Option is required:
- NAME - Uniq name for new event. Should be one word with chars and digits only

`/event` my _NAME_ rm

Remove your personal bet from the particular event

Option is required:
- NAME - Uniq name for new event. Should be one word with chars and digits only

`/event` close _NAME_ _RESULT_

Close event with NAME and RESULT options. It could be sent in group chat or in a direct chat with Valera.

You should have admin rights.

Options are required:
- _NAME_ - Uniq name for existing event. Should be one word with chars and digits only
- _RESULT_ - Result of the event. Should be number

`/event` result _NAME_

Show result for event with given name. It could be sent in group chat or in a direct chat with Valera.

Option is required:
- _NAME_ - Uniq name for existing event. Should be one word with chars and digits only

`/event` bet _NAME_ _VALUE_

Make your bet with value. It could be sent in group chat or in a direct chat with Valera.

Options are required:
- _NAME_ - Uniq name for existing event. Should be one word with chars and digits only
- _VALUE_ - Your bet for this event. Should be number

`/event` share _NAME_

Share event in administered groups

Option is required:
- _NAME_ - Uniq name for existing event. Should be one word with chars and digits only`