;; recursive function definition
(defn f []
     (defn g [] 23)
     (defn h [] 77)
     (+ (g) (h)))

(assert (== (f) 100))
