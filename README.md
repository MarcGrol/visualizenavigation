# Visualize navigation

Convert user navigation logs (timestamp, session-id, screen-name) as input to build a graph of user navigation.
Grapviz is used for visualisation.

##
    # Assume you have golang installed
    
    # install the tool
    git clone https://github.com/MarcGrol/visualizenavigation.git
    cd visualizenavigation
    go install
    
    # run the tool
    visualizenavigation -input-filename=logs.csv -limit=100 > logs.dot 
    
    # Use graphviz to create on image
    dot -Tpng -o logs.png logs.dot


## Result

Based on [this CVS file](https://github.com/MarcGrol/visualizenavigation/blob/4b8c79ad840c6ca7987581db0ef2b472a7a87b85/logs.csv) the following output is created

![example](https://github.com/MarcGrol/visualizenavigation/blob/e52527a3a844cbd5cdf56a724a71898ade7a9713/logs.png)



