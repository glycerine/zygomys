(defn add3 [a] (+ a 3))

(assert (== '(4 5 6) (map add3 '(1 2 3))))
(assert (== [4 5 6] (map add3 [1 2 3])))

(assert (== (apply (fn [a b] (* 2 a b)) [1 2]) 4))
(assert (== (apply (fn [a b] (* 2 a b)) '(1 2)) 4))
