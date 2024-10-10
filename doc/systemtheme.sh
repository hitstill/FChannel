#!/bin/sh
#Basic script to merge default and tomorrow theme into system theme

# Light theme
printf "@media (prefers-color-scheme: light) {\n" > views/css/themes/system.css
cat views/css/themes/default.css >> views/css/themes/system.css
printf "\n}" >> views/css/themes/system.css

# Dark theme
printf "\n\n@media (prefers-color-scheme: dark) {\n" >> views/css/themes/system.css
cat views/css/themes/tomorrow.css >> views/css/themes/system.css
printf "\n}" >> views/css/themes/system.css