# navigation

Linear activation history for back and forward across editor groups. Visits
are recorded on activation, forward entries are discarded on a new visit,
consecutive duplicates collapse, closed views are forgotten and the index is
repaired. The workbench store owns applying it to view activation.
