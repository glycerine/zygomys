(let* [a 0
       b a]
  (assert (= b 0)))

(let [a 5]
  (let []
    (def a 6)
    (assert (= a 6)))
  (assert (= a 5)))
