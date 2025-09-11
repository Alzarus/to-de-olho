import { useState } from 'react';
import { User, Building2, MapPin, Euro } from 'lucide-react';
import { Deputado } from './DeputadosPage';

interface DeputadoCardProps {
  deputado: Deputado;
  onClick: () => void;
}

export default function DeputadoCard({ deputado, onClick }: DeputadoCardProps) {
  const [imageError, setImageError] = useState(false);

  return (
    <div
      className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
    >
      <div className="flex items-center mb-4">
        {deputado.urlFoto && !imageError ? (
          <img
            src={deputado.urlFoto}
            alt={deputado.nome}
            className="w-16 h-16 rounded-full object-cover mr-4"
            onError={() => setImageError(true)}
          />
        ) : (
          <div className="w-16 h-16 bg-gray-200 rounded-full flex items-center justify-center mr-4">
            <User className="h-8 w-8 text-gray-400" />
          </div>
        )}
        <div>
          <h3 className="font-semibold text-gray-900 text-sm leading-tight">
            {deputado.nome}
          </h3>
          <div className="flex items-center text-sm text-gray-600 mt-1">
            <Building2 className="h-3 w-3 mr-1" />
            {deputado.siglaPartido}
            <MapPin className="h-3 w-3 ml-2 mr-1" />
            {deputado.siglaUf}
          </div>
        </div>
      </div>
      
      <div className="text-sm text-gray-600">
        <p><strong>Situação:</strong> {deputado.condicaoEleitoral}</p>
        {deputado.email && (
          <p className="mt-1 truncate"><strong>Email:</strong> {deputado.email}</p>
        )}
      </div>

      <button className="mt-4 w-full bg-blue-50 text-blue-700 py-2 px-4 rounded-md hover:bg-blue-100 transition-colors flex items-center justify-center text-sm font-medium">
        <Euro className="h-4 w-4 mr-2" />
        Ver Despesas
      </button>
    </div>
  );
}
