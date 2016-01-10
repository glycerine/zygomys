(defmac h [a] (list 4 a a a a))
(def x (macexpand (h 78)))
(assert (== (str x) "(quote 4 78 78 78 78)"))
