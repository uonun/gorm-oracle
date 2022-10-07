-- Create table
create table CUSTOMERS
(
  customer_id   NUMBER(10) not null,
  customer_name VARCHAR2(200) not null,
  address       VARCHAR2(500),
  city          VARCHAR2(500),
  state         VARCHAR2(100),
  zip_code      VARCHAR2(100),
  created_time          Date,
  age number ,
  CONSTRAINT customer_pk PRIMARY KEY (customer_id)
)
tablespace USERS
  pctfree 10
  initrans 1
  maxtrans 255;
-- Add comments to the columns 
comment on column CUSTOMERS.customer_name
  is 'customer name';
-- Create Sequence
CREATE SEQUENCE CUSTOMERS_s
minvalue 1
maxvalue 999999999999999999999999999
start with 1
increment by 1
cache 20;
------------------------------------------------------------------------------------------------
-- Create table
create table DEPARTMENTS
(
  department_id   NUMBER(10) not null,
  department_name VARCHAR2(100) not null,
  CONSTRAINT department_pk PRIMARY KEY (department_id)
)
tablespace USERS
  pctfree 10
  initrans 1
  maxtrans 255;
-- Add comments to the columns 
comment on column DEPARTMENTS.department_name
  is 'department name';
-- Create Sequence
CREATE SEQUENCE DEPARTMENTS_s
minvalue 1
maxvalue 999999999999999999999999999
start with 1
increment by 1
cache 20;
------------------------------------------------------------------------------------------------
-- Create table
create table EMPLOYEES
(
  employee_id NUMBER(10) not null,
  employee_name   VARCHAR2(100) not null,
  department_id   NUMBER(10),
  salary          NUMBER(6),
  CONSTRAINT employee_pk PRIMARY KEY (employee_id),
  constraint FK_DEPARTMENTS foreign key (DEPARTMENT_ID)
  references DEPARTMENTS (DEPARTMENT_ID)
)
tablespace USERS
  pctfree 10
  initrans 1
  maxtrans 255;
-- Add comments to the columns 
comment on column EMPLOYEES.employee_name
  is 'employee name';
-- Create Sequence
CREATE SEQUENCE EMPLOYEES_s
minvalue 1
maxvalue 999999999999999999999999999
start with 1
increment by 1
cache 20;