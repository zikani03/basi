# Actions

Actions are a way to instruct the browser on what to do on the page, this ranges from filling forms, uploading files, Clicking buttons and Page navigation. Most activities you can perform via a browser are presented as actions.

Most actions require an element or part of a page to act on and some actions require additional parameters like what data to input into a form or what file to upload.


| Action                | Arguments                        | Example                                                                                       |
| --------------------- | -------------------------------- | --------------------------------------------------------------------------------------------- |
| Click                 | **querySelector**                | Click "#element"                                                                              |
| DoubleClick           | **querySelector**                | DoubleClick "#element"                                                                        |
| Tap                   | **querySelector**                | Tap "#element"                                                                                |
| Focus                 | **querySelector**                | Focus "#element"                                                                              |
| Blur                  | **querySelector**                | Blur "#element"                                                                               |
| Fill                  | **querySelector** TEXT           | Fill "#element" "my input text"                                                               |
| Find                  | **textContent or querySelector** | Find "My Account"                                                                             |
| Clear                 | **querySelector**                | Clear "#element"                                                                              |
| Check                 | **querySelector**                | Check "#element"                                                                              |
| Uncheck               | **querySelector**                | Uncheck "#element"                                                                            |
| FillCheckbox          | **querySelector**                | FillCheckbox "#element"                                                                       |
| Press                 | **querySelector** TEXT           | Press "#element" "some text"                                                                  |
| PressSequentially     | **querySelector** TEXT           | PressSequentially "#element" "some input"                                                     |
| Type                  | **querySelector** TEXT           | Type "#element"                                                                               |
| Screenshot            | **querySelector** TEXT           | Screenshot "#selector" "filename.png"                                                         |
| Select                | **querySelector** TEXT           | Select "#someSelect" "Value or Label"                                                         |
| SelectOption          | **querySelector** TEXT           | Select "#someSelect" "Value or Label"                                                         |
| SelectMultipleOptions | **querySelector** TEXT           | SelectMultipleOptions "#someSelect" "Value or Label 1,Value or Label 2,..., Value or Label N" |
| WaitFor               | **querySelector**                | WaitFor "#element"                                                                            |
| WaitForSelector       | **querySelector**                | WaitForSelector "#element"                                                                    |
| Goto                  | **REGEX**                        | Goto "^some-page"                                                                             |
| WaitForURL            | **REGEX**                        | WaitForURL "^some-page"                                                                       |
| GoBack                | N/A                              | GoBack                                                                                        |
| GoForward             | N/A                              | GoForward                                                                                     |
| Refresh               | N/A                              | Refresh                                                                                       |

