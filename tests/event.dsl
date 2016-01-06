(defmap event)
(defmap person)
(def ev1 (event id:456 user: (person first:"jay" last:"son") flight:"A" pilot:["u" "2"]))

;; togo has the side effect of decoding the Sexp into a predefined go struct.
;; The only advance requirement is that the struct be added to interpreter/makego.go's
;; registry.
(togo ev1)

;; verify it is there with
;; (dump ev1)

;;(assert (== (ev1.DisplayEvent "prefix") `prefix gdsl.Event{Id:456, User:gdsl.Person{First:"jay", Last:"son"}, Flight:"A", Pilot:[]string{"u", "2"}}'`))


