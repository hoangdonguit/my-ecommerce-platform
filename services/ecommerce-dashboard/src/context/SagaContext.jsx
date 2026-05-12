import React, { createContext, useState, useContext } from 'react';

const SagaContext = createContext();

export const SagaProvider = ({ children }) => {
  const [products, setProducts] = useState([]);
  const [orders, setOrders] = useState([]);
  const [servicesHealth, setServicesHealth] = useState(null);

  return (
    <SagaContext.Provider value={{ 
      products, setProducts, 
      orders, setOrders, 
      servicesHealth, setServicesHealth 
    }}>
      {children}
    </SagaContext.Provider>
  );
};

export const useSaga = () => useContext(SagaContext);