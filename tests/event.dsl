(def ev1 (event id:456 user: (person first:"jay" last:"son") flight:"A" pilot:["u" "2"]))

(assert (== (ev1.DisplayEvent "prefix") `prefix gdsl.Event{Id:456, User:gdsl.Person{First:"jay", Last:"son"}, Flight:"A", Pilot:[]string{"u", "2"}}'`))


