package zygo

import (
	"time"
)

// this one specific file, calendar.go, has some portions derived
// from quantlib, that are used under the following license.
/*
 Copyright (C) 2004, 2005 Ferdinando Ametrano
 Copyright (C) 2000, 2001, 2002, 2003 RiskMap srl
 Copyright (C) 2003, 2004, 2005, 2006 StatPro Italia srl
 Copyright (C) 2017 Peter Caspers
 Copyright (C) 2017 Oleg Kulkov

 This file is part of QuantLib, a free-software/open-source library
 for financial quantitative analysts and developers - http://quantlib.org/

 QuantLib is free software: you can redistribute it and/or modify it
 under the terms of the QuantLib license.  You should have received a
 copy of the license along with this program; if not, please email
 <quantlib-dev@lists.sf.net>. The license is also available online at
 <http://quantlib.org/license.shtml>.

 This program is distributed in the hope that it will be useful, but WITHOUT
 ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 FOR A PARTICULAR PURPOSE.  See the license for more details.

LICENSE:

QuantLib is
    Copyright (C) 2000, 2001, 2002, 2003 RiskMap srl

    Copyright (C) 2001, 2002, 2003 Nicolas Di Césaré
    Copyright (C) 2001, 2002, 2003 Sadruddin Rejeb

    Copyright (C) 2002, 2003, 2004 Decillion Pty(Ltd)
    Copyright (C) 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2014, 2015 Ferdinando Ametrano

    Copyright (C) 2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2014, 2016, 2017 StatPro Italia srl
    Copyright (C) 2003, 2004, 2007 Neil Firth
    Copyright (C) 2003, 2004 Roman Gitlin
    Copyright (C) 2003 Niels Elken Sønderby
    Copyright (C) 2003 Kawanishi Tomoya

    Copyright (C) 2004 FIMAT Group
    Copyright (C) 2004 M-Dimension Consulting Inc.
    Copyright (C) 2004 Mike Parker
    Copyright (C) 2004 Walter Penschke
    Copyright (C) 2004 Gianni Piolanti
    Copyright (C) 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2013, 2014, 2015, 2016, 2017, 2018 Klaus Spanderen
    Copyright (C) 2004 Jeff Yu

    Copyright (C) 2005, 2006, 2008 Toyin Akin
    Copyright (C) 2005 Sercan Atalik
    Copyright (C) 2005, 2006 Theo Boafo
    Copyright (C) 2005, 2006, 2007, 2009 Piter Dias
    Copyright (C) 2005, 2013 Gary Kennedy
    Copyright (C) 2005, 2006, 2007 Joseph Wang
    Copyright (C) 2005 Charles Whitmore

    Copyright (C) 2006, 2007 Banca Profilo S.p.A.
    Copyright (C) 2006, 2007 Marco Bianchetti
    Copyright (C) 2006 Yiping Chen
    Copyright (C) 2006 Warren Chou
    Copyright (C) 2006, 2007 Cristina Duminuco
    Copyright (C) 2006, 2007 Giorgio Facchinetti
    Copyright (C) 2006, 2007 Chiara Fornarola
    Copyright (C) 2006 Silvia Frasson
    Copyright (C) 2006 Richard Gould
    Copyright (C) 2006, 2007, 2008, 2009, 2010 Mark Joshi
    Copyright (C) 2006, 2007, 2008 Allen Kuo
    Copyright (C) 2006, 2007, 2008, 2009, 2012 Roland Lichters
    Copyright (C) 2006, 2007 Katiuscia Manzoni
    Copyright (C) 2006, 2007 Mario Pucci
    Copyright (C) 2006, 2007 François du Vignaud

    Copyright (C) 2007 Affine Group Limited
    Copyright (C) 2007 Richard Gomes
    Copyright (C) 2007, 2008 Laurent Hoffmann
    Copyright (C) 2007, 2008, 2009, 2010, 2011 Chris Kenyon
    Copyright (C) 2007 Gang Liang

    Copyright (C) 2008, 2009, 2014, 2015, 2016 Jose Aparicio
    Copyright (C) 2008 Yee Man Chan
    Copyright (C) 2008, 2011 Charles Chongseok Hyun
    Copyright (C) 2008 Piero Del Boca
    Copyright (C) 2008 Paul Farrington
    Copyright (C) 2008 Lorella Fatone
    Copyright (C) 2008, 2009 Andreas Gaida
    Copyright (C) 2008 Marek Glowacki
    Copyright (C) 2008 Florent Grenier
    Copyright (C) 2008 Frank Hövermann
    Copyright (C) 2008 Simon Ibbotson
    Copyright (C) 2008 John Maiden
    Copyright (C) 2008 Francesca Mariani
    Copyright (C) 2008, 2009, 2010, 2011, 2012, 2014 Master IMAFA - Polytech'Nice Sophia - Université de Nice Sophia Antipolis
    Copyright (C) 2008, 2009 Andrea Odetti
    Copyright (C) 2008 J. Erik Radmall
    Copyright (C) 2008 Maria Cristina Recchioni
    Copyright (C) 2008, 2009, 2012, 2014 Ralph Schreyer
    Copyright (C) 2008 Roland Stamm
    Copyright (C) 2008 Francesco Zirilli

    Copyright (C) 2009 Nathan Abbott
    Copyright (C) 2009 Sylvain Bertrand
    Copyright (C) 2009 Frédéric Degraeve
    Copyright (C) 2009 Dirk Eddelbuettel
    Copyright (C) 2009 Bernd Engelmann
    Copyright (C) 2009, 2010, 2012 Liquidnet Holdings, Inc.
    Copyright (C) 2009 Bojan Nikolic
    Copyright (C) 2009, 2010 Dimitri Reiswich
    Copyright (C) 2009 Sun Xiuxin

    Copyright (C) 2010 Kakhkhor Abdijalilov
    Copyright (C) 2010 Hachemi Benyahia
    Copyright (C) 2010 Manas Bhatt
    Copyright (C) 2010 DeriveXperts SAS
    Copyright (C) 2010, 2014 Cavit Hafizoglu
    Copyright (C) 2010 Michael Heckl
    Copyright (C) 2010 Slava Mazur
    Copyright (C) 2010, 2011, 2012, 2013 Andre Miemiec
    Copyright (C) 2010 Adrian O' Neill
    Copyright (C) 2010 Robert Philipp
    Copyright (C) 2010 Alessandro Roveda
    Copyright (C) 2010 SunTrust Bank

    Copyright (C) 2011, 2013, 2014 Fabien Le Floc'h

    Copyright (C) 2012, 2013 Grzegorz Andruszkiewicz
    Copyright (C) 2012, 2013, 2014, 2015, 2016, 2017 Peter Caspers
    Copyright (C) 2012 Mateusz Kapturski
    Copyright (C) 2012 Simon Shakeshaft
    Copyright (C) 2012 Édouard Tallent
    Copyright (C) 2012 Samuel Tebege

    Copyright (C) 2013 BGC Partners L.P.
    Copyright (C) 2013 Chris Higgs
    Copyright (C) 2013, 2014, 2015 Cheng Li
    Copyright (C) 2013 Yue Tian

    Copyright (C) 2014, 2017 Francois Botha
    Copyright (C) 2014, 2015 Johannes Goettker-Schnetmann
    Copyright (C) 2014 Michal Kaut
    Copyright (C) 2014, 2015 Bernd Lewerenz
    Copyright (C) 2014, 2015, 2016 Paolo Mazzocchi
    Copyright (C) 2014, 2015 Thema Consulting SA
    Copyright (C) 2014, 2015, 2016 Michael von den Driesch

    Copyright (C) 2015 Riccardo Barone
    Copyright (C) 2015 CompatibL
    Copyright (C) 2015, 2016 Andres Hernandez
    Copyright (C) 2015 Dmitri Nesteruk
    Copyright (C) 2015 Maddalena Zanzi

    Copyright (C) 2016 Nicholas Bertocchi
    Copyright (C) 2016 Stefano Fondi
    Copyright (C) 2016, 2017 Fabrice Lecuyer
    Copyright (C) 2016 Eisuke Tani

    Copyright (C) 2017 BN Algorithms Ltd
    Copyright (C) 2017 Paul Giltinan
    Copyright (C) 2017 Werner Kuerzinger
    Copyright (C) 2017 Oleg Kulkov
    Copyright (C) 2017 Joseph Jeisman

    Copyright (C) 2018 Roy Zywina

QuantLib includes code taken from Peter Jäckel's book "Monte Carlo
Methods in Finance".

QuantLib includes software developed by the University of Chicago,
as Operator of Argonne National Laboratory.

QuantLib includes a set of numbers provided by Stephen Joe and Frances
Kuo under a BSD-style license.


Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

    Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

    Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

    Neither the names of the copyright holders nor the names of the QuantLib
    Group and its contributors may be used to endorse or promote products
    derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDERS OR
CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF
USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED
AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.

*/

const (
	January   = 1
	February  = 2
	March     = 3
	April     = 4
	May       = 5
	June      = 6
	July      = 7
	August    = 8
	September = 9
	October   = 10
	November  = 11
	December  = 12
	Jan       = 1
	Feb       = 2
	Mar       = 3
	Apr       = 4
	Jun       = 6
	Jul       = 7
	Aug       = 8
	Sep       = 9
	Oct       = 10
	Nov       = 11
	Dec       = 12
)

func isWashingtonBirthday(d int, m int, y int, w time.Weekday) bool {

	if y >= 1971 {
		// third Monday in February
		return (d >= 15 && d <= 21) && w == time.Monday && m == February
	} else {
		// February 22nd, possily adjusted
		return (d == 22 || (d == 23 && w == time.Monday) ||
			(d == 21 && w == time.Friday)) && m == February
	}
}

func isMemorialDay(d int, m int, y int, w time.Weekday) bool {

	if y >= 1971 {
		// last time.Monday in May
		return d >= 25 && w == time.Monday && m == May
	} else {
		// May 30th, possibly adjusted
		return (d == 30 || (d == 31 && w == time.Monday) ||
			(d == 29 && w == time.Friday)) && m == May
	}
}

func isLaborDay(d int, m int, y int, w time.Weekday) bool {

	// first time.Monday in September
	return d <= 7 && w == time.Monday && m == September
}

func isColumbusDay(d int, m int, y int, w time.Weekday) bool {

	// second time.Monday in October
	return (d >= 8 && d <= 14) && w == time.Monday && m == October &&
		y >= 1971
}

func isVeteransDay(d int, m int, y int, w time.Weekday) bool {

	if y <= 1970 || y >= 1978 {
		// November 11th, adjusted
		return (d == 11 || (d == 12 && w == time.Monday) ||
			(d == 10 && w == time.Friday)) && m == November
	} else {
		// fourth time.Monday in October
		return (d >= 22 && d <= 28) && w == time.Monday && m == October
	}
}

func isVeteransDayNoSaturday(d int, m int, y int, w time.Weekday) bool {

	if y <= 1970 || y >= 1978 {
		// November 11th, adjusted, but no Saturday to time.Friday
		return (d == 11 || (d == 12 && w == time.Monday)) && m == November
	} else {
		// fourth time.Monday in October
		return (d >= 22 && d <= 28) && w == time.Monday && m == October
	}
}

var easterMondayArray = []int{
	98, 90, 103, 95, 114, 106, 91, 111, 102, // 1901-1909
	87, 107, 99, 83, 103, 95, 115, 99, 91, 111, // 1910-1919
	96, 87, 107, 92, 112, 103, 95, 108, 100, 91, // 1920-1929
	111, 96, 88, 107, 92, 112, 104, 88, 108, 100, // 1930-1939
	85, 104, 96, 116, 101, 92, 112, 97, 89, 108, // 1940-1949
	100, 85, 105, 96, 109, 101, 93, 112, 97, 89, // 1950-1959
	109, 93, 113, 105, 90, 109, 101, 86, 106, 97, // 1960-1969
	89, 102, 94, 113, 105, 90, 110, 101, 86, 106, // 1970-1979
	98, 110, 102, 94, 114, 98, 90, 110, 95, 86, // 1980-1989
	106, 91, 111, 102, 94, 107, 99, 90, 103, 95, // 1990-1999
	115, 106, 91, 111, 103, 87, 107, 99, 84, 103, // 2000-2009
	95, 115, 100, 91, 111, 96, 88, 107, 92, 112, // 2010-2019
	104, 95, 108, 100, 92, 111, 96, 88, 108, 92, // 2020-2029
	112, 104, 89, 108, 100, 85, 105, 96, 116, 101, // 2030-2039
	93, 112, 97, 89, 109, 100, 85, 105, 97, 109, // 2040-2049
	101, 93, 113, 97, 89, 109, 94, 113, 105, 90, // 2050-2059
	110, 101, 86, 106, 98, 89, 102, 94, 114, 105, // 2060-2069
	90, 110, 102, 86, 106, 98, 111, 102, 94, 114, // 2070-2079
	99, 90, 110, 95, 87, 106, 91, 111, 103, 94, // 2080-2089
	107, 99, 91, 103, 95, 115, 107, 91, 111, 103, // 2090-2099
	88, 108, 100, 85, 105, 96, 109, 101, 93, 112, // 2100-2109
	97, 89, 109, 93, 113, 105, 90, 109, 101, 86, // 2110-2119
	106, 97, 89, 102, 94, 113, 105, 90, 110, 101, // 2120-2129
	86, 106, 98, 110, 102, 94, 114, 98, 90, 110, // 2130-2139
	95, 86, 106, 91, 111, 102, 94, 107, 99, 90, // 2140-2149
	103, 95, 115, 106, 91, 111, 103, 87, 107, 99, // 2150-2159
	84, 103, 95, 115, 100, 91, 111, 96, 88, 107, // 2160-2169
	92, 112, 104, 95, 108, 100, 92, 111, 96, 88, // 2170-2179
	108, 92, 112, 104, 89, 108, 100, 85, 105, 96, // 2180-2189
	116, 101, 93, 112, 97, 89, 109, 100, 85, 105, // 2190-2199
}

func easterMonday(year int) int {
	return easterMondayArray[year-1901]
}

func IsBusinessDay(date *Date) bool {

	_, tmDate := ToUtcDate(date, UtcTz)
	dd := tmDate.YearDay()
	w := tmDate.Weekday()
	isWeekend := false
	if w == time.Saturday || w == time.Sunday {
		isWeekend = true
	}

	y := date.Year
	m := date.Month
	d := date.Day

	em := easterMonday(y)
	if isWeekend ||
		// New Year's Day (possibly moved to time.Monday if on Sunday)
		((d == 1 || (d == 2 && w == time.Monday)) && m == January) ||
		// Washington's birthday (third time.Monday in February)
		isWashingtonBirthday(d, m, y, w) ||
		// Good time.Friday
		(dd == em-3) ||
		// Memorial Day (last time.Monday in May)
		isMemorialDay(d, m, y, w) ||
		// Independence Day (time.Monday if Sunday or time.Friday if Saturday)
		((d == 4 || (d == 5 && w == time.Monday) ||
			(d == 3 && w == time.Friday)) && m == July) ||
		// Labor Day (first time.Monday in September)
		isLaborDay(d, m, y, w) ||
		// Thanksgiving Day (fourth Thursday in November)
		((d >= 22 && d <= 28) && w == time.Thursday && m == November) ||
		// Christmas (time.Monday if Sunday or time.Friday if Saturday)
		((d == 25 || (d == 26 && w == time.Monday) ||
			(d == 24 && w == time.Friday)) && m == December) {
		return false
	}
	if y >= 1998 && (d >= 15 && d <= 21) && w == time.Monday && m == January {
		// Martin Luther King's birthday (third time.Monday in January)
		return false
	}
	if (y <= 1968 || (y <= 1980 && y%4 == 0)) && m == November &&
		d <= 7 && w == time.Tuesday {
		// Presidential election days
		return false
	}

	// Special closings
	if // Hurricane Sandy
	(y == 2012 && m == October && (d == 29 || d == 30)) ||
		// President Ford's funeral
		(y == 2007 && m == January && d == 2) ||
		// President Reagan's funeral
		(y == 2004 && m == June && d == 11) ||
		// September 11-14, 2001
		(y == 2001 && m == September && (11 <= d && d <= 14)) ||
		// President Nixon's funeral
		(y == 1994 && m == April && d == 27) ||
		// Hurricane Gloria
		(y == 1985 && m == September && d == 27) ||
		// 1977 Blackout
		(y == 1977 && m == July && d == 14) ||
		// Funeral of former President Lyndon B. Johnson.
		(y == 1973 && m == January && d == 25) ||
		// Funeral of former President Harry S. Truman
		(y == 1972 && m == December && d == 28) ||
		// National Day of Participation for the lunar exploration.
		(y == 1969 && m == July && d == 21) ||
		// Funeral of former President Eisenhower.
		(y == 1969 && m == March && d == 31) ||
		// Closed all day - heavy snow.
		(y == 1969 && m == February && d == 10) ||
		// Day after Independence Day.
		(y == 1968 && m == July && d == 5) ||
		// June 12-Dec. 31, 1968
		// Four day week (closed on Wednesdays) - Paperwork Crisis
		(y == 1968 && dd >= 163 && w == time.Wednesday) ||
		// Day of mourning for Martin Luther King Jr.
		(y == 1968 && m == April && d == 9) ||
		// Funeral of President Kennedy
		(y == 1963 && m == November && d == 25) ||
		// Day before Decoration Day
		(y == 1961 && m == May && d == 29) ||
		// Day after Christmas
		(y == 1958 && m == December && d == 26) ||
		// Christmas Eve
		((y == 1954 || y == 1956 || y == 1965) &&
			m == December && d == 24) {
		return false
	}
	return true
}
