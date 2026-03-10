grammar MyGo;

// Parser Rules
program: packageDecl? importStmt* statement+ EOF;

packageDecl: 'package' ID ';'? ;

importStmt
    : 'import' '{' importSpec (',' importSpec)* ','? '}' ';'? # BlockImport
    | 'import' importSpec ';'?                                # SingleImport
    ;

importSpec
    : STRING ('as' ID)?
    ;

statement
    : varDecl
    | assignmentStmt
    | exprStmt
    | ifStmt
    | matchStmt
    | whileStmt
    | loopStmt
    | forStmt
    | breakStmt
    | continueStmt
    | structDecl
    | enumDecl
    | fnDecl
    | traitDecl
    | returnStmt
    | deferStmt
    | spawnStmt
    | selectStmt
    ;

spawnStmt: 'spawn' (block | exprStmt) ;

selectStmt: 'select' '{' selectBranch (',' selectBranch)* ','? '}' ;
selectBranch
    : selectRead '=>' block    # SelectReadBranch
    | selectWrite '=>' block   # SelectWriteBranch
    | selectOther '=>' block   # SelectOtherBranch
    ;

selectRead: 'let' (ID | '(' ID (',' ID)? ')') '=' expr '.' method=ID '(' ')' ;
selectWrite: expr '.' method=ID '(' expr ')' ;
selectOther: 'other' ;

deferStmt: 'defer' (block | exprStmt);

assignmentStmt: expr '=' expr ';' ;

modifier: 'pub' | 'pkg' | 'pri' ;

typeParams: '<' typeParam (',' typeParam)* '>' ;
typeParam: ID (':' typeType)? ('=' typeType)? ;
typeArgs: '<' typeList '>' ;
whereClause: 'where' genericConstraint (',' genericConstraint)* ;
genericConstraint: ID ':' typeType ('+' typeType)* ;

structDecl: whereClause? modifier? 'struct' ID typeParams? '{' (structField (',' structField)* ','?)? '}' ;
structField: ID ':' typeType ;

enumDecl: whereClause? modifier? 'enum' ID typeParams? '{' enumVariant (',' enumVariant)* ','? '}' ;
enumVariant: ID ('(' typeList ')')? ;

fnDecl: whereClause? modifier? 'fn' ID typeParams? '(' paramList? ')' (':' typeType)? block ;
paramList: param (',' param)* ;
param: ID ':' typeType ;

traitDecl
    : whereClause? modifier? 'trait' ID typeParams? '{' traitFnDecl* '}'                                                                 # PureTraitDecl
    | whereClause? modifier? 'trait' 'bind' typeParams? '(' bindTarget ('|' bindTarget)* ')' ('combs' '(' ID (',' ID)* ')')? '{' traitBodyItem* '}'  # BindTraitDecl
    ;

traitFnDecl
    : 'fn' ID typeParams? '(' paramList? ')' (':' typeType)? (block | ';')
    ;

bindTarget: (ID ':')? typeType ;

traitBodyItem: banDirective | traitFnDecl ;
banDirective: 'flip'? 'ban' '[' ID (',' ID)* ']' ';' # SpecificBan | 'ban' 'repeat' ';' # RepeatBan ;

returnStmt: 'return' expr? ';' ;
block: '{' statement* '}' ;

ifStmt: 'if' expr block ('else' 'if' expr block)* ('else' block)? ;
matchStmt: 'match' expr '{' matchCase+ '}' ;
matchCase
    : expr (',' expr)* '=>' (block | statement)    # ValueMatchCase
    | 'is' typeType '=>' (block | statement)       # TypeMatchCase
    | 'other' '=>' (block | statement)             # DefaultMatchCase
    ;

whileStmt: 'while' expr block ;
loopStmt: 'loop' block ;

forStmt
    : 'for' '(' ID ':' expr '..' expr ')' block                      # RangeForStmt
    | 'for' '(' forInit? ';' cond=expr? ';' step=expr? ')' block     # TraditionalForStmt
    | 'for' '(' ID (',' ID)? ':' expr ')' block                      # IteratorForStmt
    ;

forInit: 'let' ID ('=' expr)? | expr ;

breakStmt: 'break' ';' ;
continueStmt: 'continue' ';' ;

varDecl
    : modifier? 'let' ID (':' typeType)? ('=' expr)? ';'             # SingleLetDecl
    | modifier? 'let' '(' ID (',' ID)* ')' '=' expr ';'              # TupleLetDecl
    | modifier? 'const' ID (':' typeType)? '=' expr ';'              # ConstDecl
    ;

typeList: typeType (',' typeType)* ;
typeType: '*' typeType | qualifiedName typeArgs? ('[' INT? ']')? | '(' typeList ')' | 'fn' '(' typeList? ')' (':' typeType)? ;

qualifiedName: ID ('.' ID)* ;

exprStmt: expr ';' ;
exprList: expr (',' expr)* ;

expr
    : '(' paramList? ')' (':' typeType)? '=>' block         # LambdaExpr
    | '(' expr ')'                                         # ParenExpr
    | '(' expr (',' expr)+ ')'                             # TupleExpr
    | '!' expr                                             # NotExpr
    | '&' expr                                             # AddrOfExpr
    | '*' expr                                             # DerefExpr
    | expr op=('++' | '--')                                 # PostfixExpr
    | expr '.' ID typeArgs? '(' exprList? ')'                 # MethodCallExpr
    | expr '.' ID                                           # MemberAccessExpr
    | qualifiedName typeArgs? '(' exprList? ')'             # FuncCallExpr
    | expr '(' exprList? ')'                                # CallExpr
    | expr '[' expr ']'                                     # ArrayIndexExpr
    | '[' exprList? ']'                                     # ArrayLiteralExpr
    | qualifiedName typeArgs? '{' (ID ':' expr (',' ID ':' expr)* ','?)? '}' # StructLiteralExpr  // 🎯 结构体实例化 User{}
    | expr '?!' (block | statement)?                        # TryUnwrapExpr
    | expr '?!!'                                            # PanicUnwrapExpr
    | expr 'is' typeType                                    # IsExpr
    | expr '!is' typeType                                   # NotIsExpr
    | expr 'to' typeType                                    # CastExpr
    | expr op=('*'|'/') expr                                # MulDivExpr
    | expr op=('+'|'-') expr                                # AddSubExpr
    | expr op=('=='|'!='|'>'|'<'|'>='|'<=') expr                 # BinaryCompareExpr
    | expr '&&' expr                                       # LogicalAndExpr
    | expr '||' expr                                       # LogicalOrExpr
    | expr '?' expr ':' expr                                # TernaryExpr
    | 'this'                                                # ThisExpr
    | 'nil'                                                 # NilExpr
    | qualifiedName                                         # IdentifierExpr
    | INT                                                   # IntExpr
    | STRING                                                # StringExpr
    | FLOAT                                                 # FloatExpr
    ;

// Lexer Rules
// Keywords
PACKAGE: 'package';
IMPORT: 'import';
STRUCT: 'struct';
ENUM: 'enum';
FN: 'fn';
TRAIT: 'trait';
BIND: 'bind';
COMBS: 'combs';
BAN: 'ban';
FLIP: 'flip';
REPEAT: 'repeat';
WHERE: 'where';
IF: 'if';
ELSE: 'else';
MATCH: 'match';
IS: 'is';
OTHER: 'other';
FOR: 'for';
WHILE: 'while';
LOOP: 'loop';
BREAK: 'break';
CONTINUE: 'continue';
RETURN: 'return';
LET: 'let';
CONST: 'const';
SPAWN: 'spawn';
SELECT: 'select';
PUB: 'pub';
PRI: 'pri';
PKG: 'pkg';
TO: 'to';
THIS: 'this';
NIL: 'nil';

TRY_UNWRAP : '?!' ;
PANIC_UNWRAP: '?!!' ;
ID  : [a-zA-Z_][a-zA-Z_0-9]* ;
INT : [0-9]+ ;
FLOAT: [0-9]+ '.' [0-9]+ ;
STRING: '"' ~["]* '"' ;
LINE_COMMENT: '//' ~[\r\n]* -> skip ;
BLOCK_COMMENT: '/*' .*? '*/' -> skip ;
WS  : [ \t\r\n]+ -> skip ;
