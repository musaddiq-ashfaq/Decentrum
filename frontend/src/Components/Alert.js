import PropTypes from 'prop-types';
import React from 'react';

const Alert = ({ variant = 'default', className = '', children }) => {
  const variantClasses = {
    default: 'bg-gray-100 text-gray-800 border-gray-200',
    destructive: 'bg-red-100 text-red-800 border-red-200',
  };

  return (
    <div
      className={`border rounded-md p-3 ${variantClasses[variant]} ${className}`}
    >
      {children}
    </div>
  );
};

Alert.propTypes = {
  variant: PropTypes.oneOf(['default', 'destructive']),
  className: PropTypes.string,
  children: PropTypes.node.isRequired,
};

const AlertDescription = ({ children }) => {
  return <div className="text-sm">{children}</div>;
};

AlertDescription.propTypes = {
  children: PropTypes.node.isRequired,
};

export { Alert, AlertDescription };
