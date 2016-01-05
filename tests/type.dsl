(defn a [] "hi")
(assert (== (type a) "function"))
(assert (== (type 123) "int"))
(assert (== (type 5.5) "float64"))
(assert (== (type {}) "hash"))
(assert (== (type []) "array"))
(assert (== (type ()) "null"))
(assert (== (type '('a . 'b)) "list"))


