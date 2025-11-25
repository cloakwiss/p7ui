package p7

// This is copied from ntquery package
import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
)

func OpenDB() (*sql.DB, func() error) {
	cmd := exec.Command("go", "env", "GOMOD")
	output, err := cmd.Output()
	if err != nil {
		log.Panicln("Find read GOMOD :", err)
	}
	dir := string(output[:bytes.LastIndexByte(output, byte('/'))])
	fullpath := fmt.Sprintf("%s/%s", dir, "ntdocs.db")

	db, err := sql.Open("sqlite3", fullpath)
	if err != nil {
		log.Panicln("Cannot open ntdocs.db :", err)
	}
	return db, db.Close
}

type Search struct {
	dbconnection *sql.DB
	cache        map[string]FunctionData
	// this one does not work for
	size uint
}

// TODO: This connection needs to be closed properly
// TODO: Cache size restriction is not implemented for now
func NewSearch(connection *sql.DB, cacheSize uint) Search {
	return Search{
		connection,
		make(map[string]FunctionData),
		cacheSize,
	}
}

func Get(s *Search, function_name string) FunctionData {
	if data, found := s.cache[function_name]; found {
		return data
	}
	data := query(s.dbconnection, function_name)
	s.cache[function_name] = data
	return data
}

// This function interacts with the database for query and should not be called directly
func query(dbConnection *sql.DB, function_name string) FunctionData {
	functionSymbols, er := dbConnection.Prepare(`SELECT FunctionSymbols.name, FunctionSymbols.arity, FunctionSymbols.return, FunctionSymbols.description, FunctionSymbols.requirements
		FROM FunctionSymbols WHERE FunctionSymbols.name = ?;`)
	if er != nil {
		log.Panicf("Failed to prepare the FunctionSymbols query, due to: %v", er)
	}
	defer functionSymbols.Close()

	functionParameters, er := dbConnection.Prepare(`SELECT FunctionParameters.srno, FunctionParameters.name, FunctionParameters.datatype, FunctionParameters.usage, FunctionParameters.documentation
		FROM FunctionParameters WHERE FunctionParameters.function_name = ? AND FunctionParameters.srno <= ? ORDER BY FunctionParameters.srno;`)
	if er != nil {
		log.Panic("Failed to prepare the FunctionParameter query, due to: %+w", er)
	}
	defer functionParameters.Close()

	var functionData FunctionData
	{
		resultingSymbol, er := functionSymbols.Query(function_name)
		if er != nil {
			log.Panicf("Query of Function Symbols table failed due to: %v", er)
		}
		defer resultingSymbol.Close()

		if resultingSymbol.Next() {
			if er := resultingSymbol.Scan(&functionData.Name, &functionData.Arity, &functionData.Return, &functionData.Description, &functionData.Requirement); er != nil {
				log.Panicf("Some error %v while scanning %s's result \n", er, "FunctionSymbol")
			}
		}
		if resultingSymbol.Next() {
			log.Panic("Only 1 Row is expected from the Function Symbols table")
		}
		if functionData.Name != function_name {
			log.Panicln("Search failed!!!!")
		}
	}
	{
		resultingParameters, er := functionParameters.Query(functionData.Name, functionData.Arity)
		if er != nil {
			log.Panic("Query of Function Parameter table failed due to: %+w", er)
		}
		defer resultingParameters.Close()

		functionData.FunctionParameters = make(FunctionParameters, 0, int(functionData.Arity))
		for i := 0; i < int(functionData.Arity) && resultingParameters.Next(); i += 1 {
			var functionPara FunctionParameter
			var num int
			if er := resultingParameters.Scan(&num, &functionPara.Name, &functionPara.Datatype, &functionPara.Usage, &functionPara.Documentation); er != nil {
				log.Panicf("Some error %v while scanning %s's result \n", er, "FunctionParameter")
			}
			// This should not be possible
			if num != i+1 {
				log.Panicln("Out of order")
			}
			functionData.FunctionParameters = append(functionData.FunctionParameters, functionPara)
		}
		// This should not be possible
		if resultingParameters.Next() {
			log.Panicf("Expected only %d rows in result, found more.\n", functionData.Arity)
		}
	}
	return functionData
}

type FunctionData struct {
	Name, Return, Description, Requirement string
	Arity                                  uint
	FunctionParameters
}

type FunctionParameters []FunctionParameter

type FunctionParameter struct {
	Name, Datatype, Usage, Documentation string
}
