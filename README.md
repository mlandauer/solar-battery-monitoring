# Solar battery monitoring

The aim of this project is to monitor an off-grid domestic battery solar setup to get regular updates
and store historical data to understand power usage and performance of the whole setup.

We'll be collecting data from a [Plastmatronic](http://www.plasmatronics.com.au/) PL80 regulator. This is a battery charger controller. So, it knows the power used by the house. It knows how charged the batteries are and how much they're being charged. It also knows the amount of power coming from the solar cells. We'll be talking to the PL80 via a PLI which is a crazy old-fashioned RS-232 serial interface.

There are a bunch of manuals which are going to be useful. We're linking to them here for easy reference:

PL80:

- [PL60/PL80 User Guide](http://www.plasmatronics.com.au/downloads/PL60.PL80.UserGuide.V6.pdf)
- [PL User Manual](http://www.plasmatronics.com.au/downloads/PLUserMan.V9.0324.pdf)
- [PL Reference Manual](http://www.plasmatronics.com.au/downloads/PL_Reference_Manual_6.3.1.pdf)

PLI:

- [PLI](http://www.plasmatronics.com.au/downloads/PLIman4.2.pdf)
- [PLI Communications](http://www.plasmatronics.com.au/downloads/PLI.Info.2.16.pdf)
