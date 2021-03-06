/************************************************************************/
/* Octatron                                                             */
/* Copyright (c) 2015 Andreas T Jonsson <mail@andreasjonsson.se>        */
/*                                                                      */
/* Octatron is free software: you can redistribute it and/or modify     */
/* it under the terms of the GNU General Public License as published by */
/* the Free Software Foundation, either version 3 of the License, or    */
/* (at your option) any later version.                                  */
/*                                                                      */
/* Octatron is distributed in the hope that it will be useful,          */
/* but WITHOUT ANY WARRANTY; without even the implied warranty of       */
/* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the        */
/* GNU General Public License for more details.                         */
/*                                                                      */
/* You should have received a copy of the GNU General Public License    */
/* along with Octatron.  If not, see <http://www.gnu.org/licenses/>.    */
/************************************************************************/

package octatron

type Sample interface {
	Color() Color
	Position() Point
}

type Worker interface {
	Run(volume Box, samples chan<- Sample) error
	Stop()
}
