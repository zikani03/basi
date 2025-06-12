Fill "x" "something something here" 
Fill "#x" "something something here"
Fill ".x" "something something here"
Fill "[name=email]" "something something here"
Fill "some element with whitespace" "something something here"
Fill ".input[type=\"text\"]" "something something here"
Click "existent" "something"
Screenshot "/path/to/file.png" "option=value, optional1=value1"
Click "GetByRole( \"existent\" )"
Click    "existent"   "this should be accepted" # "not this"
# The next line should fail
Click 
