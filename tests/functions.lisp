(assert (= 3
  ((fn [a b] a) 3 2)))

(defn add4 [a] (+ a 4))

(assert (= 7 (add4 3)))

; testing recursion
(defn fact [n]
  (cond (= n 0) 1 (* n (fact (- n 1)))))

(assert (= 120 (fact 5)))
(assert (= 3628800 (fact 10)))

; testing tail recursion
(defn fact-tc [n accum]
  (cond (= n 0) accum
    (let [newn (- n 1)
          newaccum (* accum n)]
      (fact-tc newn newaccum))))

(assert (= 120 (fact-tc 5 1)))
(assert (= 3628800 (fact-tc 10 1)))
